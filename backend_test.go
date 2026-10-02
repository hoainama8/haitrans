package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExtractWordParagraphsSkipsWhitespaceAndDecodesEntities(t *testing.T) {
	document := []byte(`<w:document><w:p><w:r><w:t>Hello</w:t></w:r><w:r><w:t xml:space="preserve"> &amp; world</w:t></w:r></w:p><w:p><w:r><w:t>&#32;</w:t></w:r></w:p><w:p><w:r><w:t>Next</w:t></w:r></w:p></w:document>`)

	paragraphs, err := extractWordParagraphs(document, 0)
	if err != nil {
		t.Fatalf("extractWordParagraphs() error = %v", err)
	}
	if len(paragraphs) != 2 {
		t.Fatalf("got %d paragraphs, want 2", len(paragraphs))
	}
	if paragraphs[0].ID != 0 || paragraphs[0].Text != "Hello & world" {
		t.Fatalf("first paragraph = %#v", paragraphs[0])
	}
	if paragraphs[1].ID != 1 || paragraphs[1].Text != "Next" {
		t.Fatalf("second paragraph = %#v", paragraphs[1])
	}
}

func TestClassifyFileVerifiesExtensionAndContents(t *testing.T) {
	kind, err := classifyFile("notes.txt", []byte("xin chào"))
	if err != nil || kind != fileTypeText {
		t.Fatalf("classify text = %q, %v", kind, err)
	}

	if _, err := classifyFile("renamed.pdf", []byte("not a PDF")); err == nil {
		t.Fatal("expected renamed non-PDF file to be rejected")
	}

	var docx bytes.Buffer
	writer := zip.NewWriter(&docx)
	file, err := writer.Create("[Content_Types].xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte("<Types/>")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	kind, err = classifyFile("document.docx", docx.Bytes())
	if err != nil || kind != fileTypeDOCX {
		t.Fatalf("classify DOCX = %q, %v", kind, err)
	}
}

func TestReplaceWordParagraphsPreservesImagesAndEquations(t *testing.T) {
	document := []byte(`<w:document><w:p><w:r><w:t>Source</w:t></w:r><w:r><w:drawing><image/></w:drawing></w:r></w:p><w:p><m:oMath><m:r><m:t>x=2</m:t></m:r></m:oMath></w:p><w:p><w:r><w:t>&#32;</w:t></w:r></w:p><w:p><w:r><w:t>Code: </w:t></w:r><w:r><w:t>fmt.Println(1)</w:t></w:r></w:p></w:document>`)
	id := 0
	updated, err := replaceWordParagraphs(document, map[int]string{
		0: "Texte traduit",
		1: "Code : fmt.Println(1)",
	}, &id)
	if err != nil {
		t.Fatalf("replaceWordParagraphs() error = %v", err)
	}
	if id != 2 {
		t.Fatalf("advanced paragraph ID = %d, want 2", id)
	}
	for _, retained := range []string{
		`<w:drawing><image/></w:drawing>`,
		`<m:oMath><m:r><m:t>x=2</m:t></m:r></m:oMath>`,
		`<w:t>Texte traduit</w:t>`,
		`<w:t>Code : fmt.Println(1)</w:t>`,
	} {
		if !strings.Contains(string(updated), retained) {
			t.Errorf("updated DOCX XML does not contain %q", retained)
		}
	}
	if strings.Contains(string(updated), "<w:t>Source</w:t>") {
		t.Error("original prose was not replaced")
	}
}

func TestWriteTranslatedDocxPreservesEmbeddedAssets(t *testing.T) {
	sourceXML := []byte(`<w:document><w:p><w:r><w:t>Source</w:t></w:r></w:p></w:document>`)
	image := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a}
	parts := []documentZipPart{
		{header: zip.FileHeader{Name: "word/document.xml", Method: zip.Deflate}, data: sourceXML},
		{header: zip.FileHeader{Name: "word/media/image1.png", Method: zip.Store}, data: image},
	}

	packed, translated, err := writeTranslatedDocx(parts, map[int]string{0: "Translated"}, 1)
	if err != nil {
		t.Fatalf("writeTranslatedDocx() error = %v", err)
	}
	if translated != "Translated" {
		t.Fatalf("translated text = %q, want %q", translated, "Translated")
	}
	reader, err := zip.NewReader(bytes.NewReader(packed), int64(len(packed)))
	if err != nil {
		t.Fatalf("output DOCX is invalid: %v", err)
	}
	for _, file := range reader.File {
		if file.Name == "word/media/image1.png" {
			asset, err := file.Open()
			if err != nil {
				t.Fatalf("open preserved image: %v", err)
			}
			var actual bytes.Buffer
			_, readErr := actual.ReadFrom(asset)
			closeErr := asset.Close()
			if readErr != nil {
				t.Fatalf("read preserved image: %v", readErr)
			}
			if closeErr != nil {
				t.Fatalf("close preserved image: %v", closeErr)
			}
			if !bytes.Equal(actual.Bytes(), image) {
				t.Fatalf("preserved image bytes = %v, want %v", actual.Bytes(), image)
			}
			return
		}
	}
	t.Fatal("embedded image was not preserved")
}

func TestTranslatedDocxExportsAsPDF(t *testing.T) {
	if _, err := exec.LookPath("libreoffice"); err != nil {
		if _, err := exec.LookPath("soffice"); err != nil {
			t.Skip("LibreOffice is not installed")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	documentXML := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body><w:p><w:r><w:t>Source text</w:t></w:r></w:p><w:sectPr/></w:body></w:document>`)
	parts := []documentZipPart{
		{header: zip.FileHeader{Name: "[Content_Types].xml", Method: zip.Deflate}, data: []byte(`<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`)},
		{header: zip.FileHeader{Name: "_rels/.rels", Method: zip.Deflate}, data: []byte(`<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`)},
		{header: zip.FileHeader{Name: "word/document.xml", Method: zip.Deflate}, data: documentXML},
	}
	docx, translated, err := writeTranslatedDocx(parts, map[int]string{0: "Translated text"}, 1)
	if err != nil {
		t.Fatalf("writeTranslatedDocx() error = %v", err)
	}
	if translated != "Translated text" {
		t.Fatalf("translated text = %q", translated)
	}

	tempDir := t.TempDir()
	docxPath := filepath.Join(tempDir, "translated.docx")
	if err := os.WriteFile(docxPath, docx, 0o600); err != nil {
		t.Fatalf("write translated DOCX: %v", err)
	}
	if err := officeConvert(ctx, tempDir, docxPath, "pdf"); err != nil {
		t.Fatalf("export translated DOCX to PDF: %v", err)
	}
	pdf, err := os.ReadFile(filepath.Join(tempDir, "translated.pdf"))
	if err != nil {
		t.Fatalf("read exported PDF: %v", err)
	}
	if len(pdf) < 5 || string(pdf[:5]) != "%PDF-" {
		t.Fatal("DOCX export did not produce a PDF")
	}
}

func TestLocalBackendRejectsRequestsWithoutToken(t *testing.T) {
	backend, err := startLocalBackend()
	if err != nil {
		t.Fatalf("startLocalBackend() error = %v", err)
	}
	defer backend.close()

	request, err := http.NewRequest(http.MethodPost, backend.url, strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("backend request error = %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusUnauthorized)
	}
}

func TestCallGeminiUsesAPIKeyHeaderAndJSONResponseMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("request method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/v1beta/interactions" {
			t.Errorf("request path = %q", r.URL.Path)
		}
		if r.Header.Get("x-goog-api-key") != "test-key" {
			t.Errorf("Gemini API key was not sent in the expected header")
		}
		var request struct {
			Model          string          `json:"model"`
			Input          json.RawMessage `json:"input"`
			Store          bool            `json:"store"`
			ResponseFormat json.RawMessage `json:"response_format"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode Interactions request: %v", err)
		}
		if request.Model != "gemini-3.8-flash" {
			t.Errorf("model = %q", request.Model)
		}
		for _, expected := range []string{`"type":"text"`, `"text":"prompt"`} {
			if !bytes.Contains(request.Input, []byte(expected)) {
				t.Errorf("input does not contain %q: %s", expected, request.Input)
			}
		}
		if request.Store {
			t.Error("expected interaction storage to be disabled")
		}
		if !bytes.Contains(request.ResponseFormat, []byte(`"application/json"`)) {
			t.Errorf("response_format does not request JSON: %s", request.ResponseFormat)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"test-interaction","status":"completed","created":"2026-09-26T00:00:00Z","updated":"2026-09-26T00:00:00Z","output_text":"translated"}`))
	}))
	defer server.Close()

	result, err := callGeminiAtEndpoint(context.Background(), server.URL, "test-key", "gemini-3.8-flash", "prompt", true)
	if err != nil {
		t.Fatalf("callGeminiAtEndpoint() error = %v", err)
	}
	if result != "translated" {
		t.Fatalf("result = %q, want %q", result, "translated")
	}
}

func TestSupportedModels(t *testing.T) {
	for _, model := range []string{"gemini-3.8-flash", "gemma-4-31b-it"} {
		if _, ok := supportedModels[model]; !ok {
			t.Errorf("model %q is not enabled in the backend", model)
		}
	}
	if _, ok := supportedModels["gemini-2.5-flash"]; ok {
		t.Error("old model should no longer be enabled")
	}
}

func TestTextToPDF(t *testing.T) {
	if _, err := exec.LookPath("libreoffice"); err != nil {
		if _, err := exec.LookPath("soffice"); err != nil {
			t.Skip("LibreOffice is not installed")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pdf, err := textToPDF(ctx, "Xin chào, đây là bản dịch.")
	if err != nil {
		t.Fatalf("textToPDF() error = %v", err)
	}
	if len(pdf) < 5 || string(pdf[:5]) != "%PDF-" {
		t.Fatal("textToPDF() did not produce a valid PDF header")
	}

	tempDir := t.TempDir()
	pdfPath := filepath.Join(tempDir, "source.pdf")
	if err := os.WriteFile(pdfPath, pdf, 0o600); err != nil {
		t.Fatalf("write test PDF: %v", err)
	}
	if err := officeConvert(ctx, tempDir, pdfPath, "fodg"); err != nil {
		t.Fatalf("convert PDF to flat ODG: %v", err)
	}
	odg, err := os.ReadFile(filepath.Join(tempDir, "source.fodg"))
	if err != nil {
		t.Fatalf("read converted flat ODG: %v", err)
	}
	paragraphs, err := extractODGParagraphs(odg, 0)
	if err != nil {
		t.Fatalf("extract imported PDF text: %v", err)
	}
	var text strings.Builder
	translations := make(map[int]string, len(paragraphs))
	for _, paragraph := range paragraphs {
		text.WriteString(paragraph.Text)
		translations[paragraph.ID] = "Bản dịch kiểm thử"
	}
	if !strings.Contains(text.String(), "Xin chào") {
		t.Fatalf("PDF-imported text = %q, want Vietnamese source text", text.String())
	}
	translatedODG, err := replaceODGParagraphs(odg, translations, new(int))
	if err != nil {
		t.Fatalf("translate imported PDF text: %v", err)
	}
	translatedODGPath := filepath.Join(tempDir, "translated.fodg")
	if err := os.WriteFile(translatedODGPath, translatedODG, 0o600); err != nil {
		t.Fatalf("write translated flat ODG: %v", err)
	}
	if err := officeConvert(ctx, tempDir, translatedODGPath, "pdf"); err != nil {
		t.Fatalf("convert translated flat ODG to PDF: %v", err)
	}
	translatedPDFPath := filepath.Join(tempDir, "translated.pdf")
	translatedPDF, err := os.ReadFile(translatedPDFPath)
	if err != nil {
		t.Fatalf("read translated PDF: %v", err)
	}
	if len(translatedPDF) < 5 || string(translatedPDF[:5]) != "%PDF-" {
		t.Fatal("translated document did not produce a valid PDF")
	}
	if pdftotext, err := exec.LookPath("pdftotext"); err == nil {
		output, err := exec.CommandContext(ctx, pdftotext, translatedPDFPath, "-").Output()
		if err != nil {
			t.Fatalf("extract text from translated PDF: %v", err)
		}
		if !strings.Contains(string(output), "Bản dịch kiểm thử") {
			t.Fatalf("translated PDF text = %q, want translated text", output)
		}
	}
}

func TestBuildTextPromptPreservesNonTranslatableContent(t *testing.T) {
	prompt := buildTextPrompt(translationRequest{
		SourceLang: "Tiếng Anh",
		TargetLang: "Tiếng Việt",
		Text:       "input",
	})
	for _, instruction := range []string{"mã nguồn", "công thức", "không dịch hoặc sửa chúng"} {
		if !strings.Contains(prompt, instruction) {
			t.Errorf("prompt does not contain instruction %q", instruction)
		}
	}
}
