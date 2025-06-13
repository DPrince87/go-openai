package openai

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
)

type FormBuilder interface {
	CreateFormFile(fieldname string, file *os.File) error
	CreateFormFileReader(fieldname string, r io.Reader, filename string) error
	CreateFormFileReaderWithMimeType(fieldname string, r io.Reader, filename string, mimeType string) error
	WriteField(fieldname, value string) error
	Close() error
	FormDataContentType() string
}

type DefaultFormBuilder struct {
	writer *multipart.Writer
}

func NewFormBuilder(body io.Writer) *DefaultFormBuilder {
	return &DefaultFormBuilder{
		writer: multipart.NewWriter(body),
	}
}

func (fb *DefaultFormBuilder) CreateFormFile(fieldname string, file *os.File) error {
	return fb.createFormFile(fieldname, file, file.Name())
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// detectMimeType attempts to detect MIME type from file content
func detectMimeType(r io.Reader) (string, io.Reader, error) {
	// Read first 512 bytes for MIME type detection
	buffer := make([]byte, 512)
	n, err := r.Read(buffer)
	if err != nil && err != io.EOF {
		return "", nil, err
	}

	// Detect MIME type
	mimeType := http.DetectContentType(buffer[:n])

	// Create a new reader that includes the read bytes
	newReader := io.MultiReader(strings.NewReader(string(buffer[:n])), r)

	return mimeType, newReader, nil
}

// CreateFormFileReader creates a form field with a file reader.
// The filename in parameters can be an empty string.
// The filename in Content-Disposition is required, But it can be an empty string.
func (fb *DefaultFormBuilder) CreateFormFileReader(fieldname string, r io.Reader, filename string) error {
	// Auto-detect MIME type if not provided
	mimeType, newReader, err := detectMimeType(r)
	if err != nil {
		return fmt.Errorf("failed to detect MIME type: %w", err)
	}

	return fb.CreateFormFileReaderWithMimeType(fieldname, newReader, filename, mimeType)
}

// CreateFormFileReaderWithMimeType creates a form field with a file reader and explicit MIME type.
func (fb *DefaultFormBuilder) CreateFormFileReaderWithMimeType(fieldname string, r io.Reader, filename string, mimeType string) error {
	// Check if the reader has a Name() method (like our NamedReader)
	if namedReader, ok := r.(interface{ Name() string }); ok && filename == "" {
		filename = namedReader.Name()
	}

	// If still no filename, provide a default based on MIME type
	if filename == "" {
		switch mimeType {
		case "image/png":
			filename = "image.png"
		case "image/jpeg", "image/jpg":
			filename = "image.jpg"
		case "image/gif":
			filename = "image.gif"
		case "image/webp":
			filename = "image.webp"
		case "image/bmp":
			filename = "image.bmp"
		case "image/tiff":
			filename = "image.tiff"
		default:
			filename = "file.bin"
		}
	}

	h := make(textproto.MIMEHeader)
	h.Set(
		"Content-Disposition",
		fmt.Sprintf(
			`form-data; name="%s"; filename="%s"`,
			escapeQuotes(fieldname),
			escapeQuotes(filepath.Base(filename)),
		),
	)

	// Set Content-Type if mimeType is provided
	if mimeType != "" {
		h.Set("Content-Type", mimeType)
	}

	fieldWriter, err := fb.writer.CreatePart(h)
	if err != nil {
		return err
	}

	_, err = io.Copy(fieldWriter, r)
	if err != nil {
		return err
	}

	return nil
}

func (fb *DefaultFormBuilder) createFormFile(fieldname string, r io.Reader, filename string) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	fieldWriter, err := fb.writer.CreateFormFile(fieldname, filename)
	if err != nil {
		return err
	}

	_, err = io.Copy(fieldWriter, r)
	if err != nil {
		return err
	}

	return nil
}

func (fb *DefaultFormBuilder) WriteField(fieldname, value string) error {
	return fb.writer.WriteField(fieldname, value)
}

func (fb *DefaultFormBuilder) Close() error {
	return fb.writer.Close()
}

func (fb *DefaultFormBuilder) FormDataContentType() string {
	return fb.writer.FormDataContentType()
}
