package openai

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

// Image sizes defined by the OpenAI API.
const (
	CreateImageSize256x256   = "256x256"
	CreateImageSize512x512   = "512x512"
	CreateImageSize1024x1024 = "1024x1024"

	// dall-e-3 supported only.
	CreateImageSize1792x1024 = "1792x1024"
	CreateImageSize1024x1792 = "1024x1792"

	// gpt-image-1 supported only.
	CreateImageSize1536x1024 = "1536x1024" // Landscape
	CreateImageSize1024x1536 = "1024x1536" // Portrait
)

const (
	// dall-e-2 and dall-e-3 only.
	CreateImageResponseFormatB64JSON = "b64_json"
	CreateImageResponseFormatURL     = "url"
)

const (
	CreateImageModelDallE2    = "dall-e-2"
	CreateImageModelDallE3    = "dall-e-3"
	CreateImageModelGptImage1 = "gpt-image-1"
)

const (
	CreateImageQualityHD       = "hd"
	CreateImageQualityStandard = "standard"

	// gpt-image-1 only.
	CreateImageQualityHigh   = "high"
	CreateImageQualityMedium = "medium"
	CreateImageQualityLow    = "low"
)

const (
	// dall-e-3 only.
	CreateImageStyleVivid   = "vivid"
	CreateImageStyleNatural = "natural"
)

const (
	// gpt-image-1 only.
	CreateImageBackgroundTransparent = "transparent"
	CreateImageBackgroundOpaque      = "opaque"
)

const (
	// gpt-image-1 only.
	CreateImageModerationLow = "low"
)

const (
	// gpt-image-1 only.
	CreateImageOutputFormatPNG  = "png"
	CreateImageOutputFormatJPEG = "jpeg"
	CreateImageOutputFormatWEBP = "webp"
)

// ImageRequest represents the request structure for the image API.
type ImageRequest struct {
	Prompt            string `json:"prompt,omitempty"`
	Model             string `json:"model,omitempty"`
	N                 int    `json:"n,omitempty"`
	Quality           string `json:"quality,omitempty"`
	Size              string `json:"size,omitempty"`
	Style             string `json:"style,omitempty"`
	ResponseFormat    string `json:"response_format,omitempty"`
	User              string `json:"user,omitempty"`
	Background        string `json:"background,omitempty"`
	Moderation        string `json:"moderation,omitempty"`
	OutputCompression int    `json:"output_compression,omitempty"`
	OutputFormat      string `json:"output_format,omitempty"`
}

// ImageResponse represents a response structure for image API.
type ImageResponse struct {
	Created int64                    `json:"created,omitempty"`
	Data    []ImageResponseDataInner `json:"data,omitempty"`
	Usage   ImageResponseUsage       `json:"usage,omitempty"`

	httpHeader
}

// ImageResponseInputTokensDetails represents the token breakdown for input tokens.
type ImageResponseInputTokensDetails struct {
	TextTokens  int `json:"text_tokens,omitempty"`
	ImageTokens int `json:"image_tokens,omitempty"`
}

// ImageResponseUsage represents the token usage information for image API.
type ImageResponseUsage struct {
	TotalTokens        int                             `json:"total_tokens,omitempty"`
	InputTokens        int                             `json:"input_tokens,omitempty"`
	OutputTokens       int                             `json:"output_tokens,omitempty"`
	InputTokensDetails ImageResponseInputTokensDetails `json:"input_tokens_details,omitempty"`
}

// ImageResponseDataInner represents a response data structure for image API.
type ImageResponseDataInner struct {
	URL           string `json:"url,omitempty"`
	B64JSON       string `json:"b64_json,omitempty"`
	RevisedPrompt string `json:"revised_prompt,omitempty"`
}

// ImageEditRequest represents the request structure for the image API.
type ImageEditRequest struct {
	Image          io.Reader `json:"image,omitempty"`
	Mask           io.Reader `json:"mask,omitempty"`
	Prompt         string    `json:"prompt,omitempty"`
	Model          string    `json:"model,omitempty"`
	N              int       `json:"n,omitempty"`
	Size           string    `json:"size,omitempty"`
	ResponseFormat string    `json:"response_format,omitempty"`
	Quality        string    `json:"quality,omitempty"`
	User           string    `json:"user,omitempty"`
}

// CreateImage - API call to create an image. This is the main endpoint of the DALL-E API.
func (c *Client) CreateImage(ctx context.Context, request ImageRequest) (response ImageResponse, err error) {
	urlSuffix := "/images/generations"
	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL(urlSuffix, withModel(request.Model)),
		withBody(request),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// ImageVariRequest represents the request structure for the image API.
type ImageVariRequest struct {
	Image          io.Reader `json:"image,omitempty"`
	Model          string    `json:"model,omitempty"`
	N              int       `json:"n,omitempty"`
	Size           string    `json:"size,omitempty"`
	ResponseFormat string    `json:"response_format,omitempty"`
	User           string    `json:"user,omitempty"`
}

// CreateVariImage - API call to create an image variation. This is the main endpoint of the DALL-E API.
// Use abbreviations(vari for variation) because ci-lint has a single-line length limit ...
func (c *Client) CreateVariImage(ctx context.Context, request ImageVariRequest) (response ImageResponse, err error) {
	body := &bytes.Buffer{}
	builder := c.createFormBuilder(body)

	// image, filename is not required
	err = builder.CreateFormFileReader("image", request.Image, "")
	if err != nil {
		return
	}

	err = builder.WriteField("n", strconv.Itoa(request.N))
	if err != nil {
		return
	}

	err = builder.WriteField("size", request.Size)
	if err != nil {
		return
	}

	err = builder.WriteField("response_format", request.ResponseFormat)
	if err != nil {
		return
	}

	err = builder.Close()
	if err != nil {
		return
	}

	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		c.fullURL("/images/variations", withModel(request.Model)),
		withBody(body),
		withContentType(builder.FormDataContentType()),
	)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}

// CreateEditImage - API call to create an image. This is the main endpoint of the DALL-E API.
func (c *Client) CreateEditImage(ctx context.Context, request ImageEditRequest) (response ImageResponse, err error) {
	// Debug logging
	fmt.Printf("[DEBUG] CreateEditImage called with: Model=%s, Prompt=%s, Size=%s, N=%d, Quality='%s', ResponseFormat='%s'\n",
		request.Model, request.Prompt, request.Size, request.N, request.Quality, request.ResponseFormat)

	body := &bytes.Buffer{}
	builder := c.createFormBuilder(body)

	// Try to get file size for debugging and ensure file is at beginning
	if file, ok := request.Image.(*os.File); ok {
		if stat, err := file.Stat(); err == nil {
			fmt.Printf("[DEBUG] Image file size: %d bytes\n", stat.Size())
		}
		// Reset file position to beginning - critical for reading
		offset, err := file.Seek(0, 0)
		fmt.Printf("[DEBUG] File seek to beginning: offset=%d, err=%v\n", offset, err)

		// Read first few bytes to verify file content
		testBytes := make([]byte, 16)
		n, err := file.Read(testBytes)
		fmt.Printf("[DEBUG] First %d bytes: %x, err=%v\n", n, testBytes[:n], err)
		// Seek back to beginning after test read
		file.Seek(0, 0)
	}

	// Check if it's a NamedReader
	if namedReader, ok := request.Image.(interface{ Name() string }); ok {
		fmt.Printf("[DEBUG] Image has filename: %s\n", namedReader.Name())
	}

	// Use CreateFormFileReader which will auto-detect MIME type and set proper filename
	err = builder.CreateFormFileReader("image", request.Image, "")
	if err != nil {
		fmt.Printf("[DEBUG] Error adding image to form: %v\n", err)
		return
	}

	// mask, it is optional
	if request.Mask != nil {
		// Try to get mask file info
		if file, ok := request.Mask.(*os.File); ok {
			if stat, err := file.Stat(); err == nil {
				fmt.Printf("[DEBUG] Mask file size: %d bytes\n", stat.Size())
			}
			// Reset file position to beginning
			offset, err := file.Seek(0, 0)
			fmt.Printf("[DEBUG] Mask file seek to beginning: offset=%d, err=%v\n", offset, err)
		}

		// Use CreateFormFileReader which will auto-detect MIME type and set proper filename
		err = builder.CreateFormFileReader("mask", request.Mask, "")
		if err != nil {
			return
		}
	}

	err = builder.WriteField("prompt", request.Prompt)
	if err != nil {
		return
	}

	// Add model field to form data
	if request.Model != "" {
		err = builder.WriteField("model", request.Model)
		if err != nil {
			return
		}
	}

	err = builder.WriteField("n", strconv.Itoa(request.N))
	if err != nil {
		return
	}

	err = builder.WriteField("size", request.Size)
	if err != nil {
		return
	}

	// CRITICAL: gpt-image-1 does NOT support response_format parameter at all
	// Python library completely omits this parameter for gpt-image-1
	if request.ResponseFormat != "" && request.Model != CreateImageModelGptImage1 {
		fmt.Printf("[DEBUG] Adding response_format: %s (model supports it)\n", request.ResponseFormat)
		err = builder.WriteField("response_format", request.ResponseFormat)
		if err != nil {
			return
		}
	} else if request.Model == CreateImageModelGptImage1 {
		fmt.Printf("[DEBUG] Skipping response_format completely for gpt-image-1 (not supported)\n")
	}

	// CRITICAL: gpt-image-1 does NOT support quality parameter in the same way
	// Python library filters this out for gpt-image-1
	if request.Quality != "" && request.Model != CreateImageModelGptImage1 {
		fmt.Printf("[DEBUG] Adding quality: %s (model supports it)\n", request.Quality)
		err = builder.WriteField("quality", request.Quality)
		if err != nil {
			return
		}
	} else if request.Quality != "" && request.Model == CreateImageModelGptImage1 {
		fmt.Printf("[DEBUG] Skipping quality for gpt-image-1 (not supported in this context)\n")
	}

	// Add user field if specified
	if request.User != "" {
		err = builder.WriteField("user", request.User)
		if err != nil {
			return
		}
	}

	err = builder.Close()
	if err != nil {
		return
	}

	// Debug: print the body size
	fmt.Printf("[DEBUG] Form body size: %d bytes\n", body.Len())

	url := c.fullURL("/images/edits")
	fmt.Printf("[DEBUG] Making request to URL: %s\n", url)
	fmt.Printf("[DEBUG] Content-Type: %s\n", builder.FormDataContentType())

	req, err := c.newRequest(
		ctx,
		http.MethodPost,
		url,
		withBody(body),
		withContentType(builder.FormDataContentType()),
	)
	if err != nil {
		fmt.Printf("[DEBUG] Error creating request: %v\n", err)
		return
	}

	fmt.Printf("[DEBUG] Sending request to OpenAI API...\n")
	fmt.Printf("[DEBUG] Request headers: %v\n", req.Header)

	err = c.sendRequest(req, &response)
	if err != nil {
		fmt.Printf("[DEBUG] Error from OpenAI API: %v\n", err)
		fmt.Printf("[DEBUG] Error type: %T\n", err)
		// Try to get more details about the error
		if apiErr, ok := err.(*APIError); ok {
			fmt.Printf("[DEBUG] API Error details - Code: %s, Message: %s, Type: %s\n",
				apiErr.Code, apiErr.Message, apiErr.Type)
		}
	} else {
		fmt.Printf("[DEBUG] Successfully received response from OpenAI API\n")
		fmt.Printf("[DEBUG] Response data count: %d\n", len(response.Data))
	}
	return
}
