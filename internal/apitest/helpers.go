package apitest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HTTPRequest represents a simplified HTTP request for testing
type HTTPRequest struct {
	Method      string
	URL         string
	Body        interface{}
	Headers     map[string]string
	ContentType string
}

// HTTPResponse represents the response from a test request
type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Response   *httptest.ResponseRecorder
}

// DoRequest executes an HTTP request against the test app
func DoRequest(t *testing.T, app *fiber.App, req HTTPRequest) *HTTPResponse {
	var body io.Reader

	// Handle body if present
	if req.Body != nil {
		var bodyBytes []byte
		var err error

		switch v := req.Body.(type) {
		case []byte:
			bodyBytes = v
		case string:
			bodyBytes = []byte(v)
		default:
			bodyBytes, err = json.Marshal(v)
			require.NoError(t, err, "Failed to marshal request body")
		}
		body = bytes.NewReader(bodyBytes)
	}

	// Create HTTP request
	httpReq := httptest.NewRequest(req.Method, req.URL, body)

	// Set content type
	contentType := req.ContentType
	if contentType == "" && req.Body != nil {
		contentType = "application/json"
	}
	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}

	// Set additional headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Execute request
	resp, err := app.Test(httpReq, -1) // -1 means no timeout
	require.NoError(t, err, "Failed to execute request")

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       bodyBytes,
		Response:   nil,
	}
}

// AssertStatusCode checks if the response status code matches expected
func AssertStatusCode(t *testing.T, resp *HTTPResponse, expectedStatus int) {
	assert.Equal(t, expectedStatus, resp.StatusCode,
		"Unexpected status code. Body: %s", string(resp.Body))
}

// AssertJSONResponse checks if response is valid JSON and unmarshals it
func AssertJSONResponse(t *testing.T, resp *HTTPResponse, target interface{}) {
	err := json.Unmarshal(resp.Body, target)
	require.NoError(t, err, "Failed to unmarshal JSON response: %s", string(resp.Body))
}

// AssertAPISuccess checks if the response is a successful API response
func AssertAPISuccess(t *testing.T, resp *HTTPResponse) map[string]interface{} {
	var result map[string]interface{}
	AssertJSONResponse(t, resp, &result)

	success, ok := result["success"].(bool)
	require.True(t, ok, "Response missing 'success' field")
	assert.True(t, success, "API response indicates failure: %v", result)

	return result
}

// AssertAPIError checks if the response is an error API response
func AssertAPIError(t *testing.T, resp *HTTPResponse) map[string]interface{} {
	var result map[string]interface{}
	AssertJSONResponse(t, resp, &result)

	success, ok := result["success"].(bool)
	require.True(t, ok, "Response missing 'success' field")
	assert.False(t, success, "API response indicates success when error expected")

	_, hasError := result["error"]
	assert.True(t, hasError, "Error response missing 'error' field")

	return result
}

// GET creates a GET request helper
func GET(url string) HTTPRequest {
	return HTTPRequest{
		Method: fiber.MethodGet,
		URL:    url,
	}
}

// POST creates a POST request helper
func POST(url string, body interface{}) HTTPRequest {
	return HTTPRequest{
		Method: fiber.MethodPost,
		URL:    url,
		Body:   body,
	}
}

// PUT creates a PUT request helper
func PUT(url string, body interface{}) HTTPRequest {
	return HTTPRequest{
		Method: fiber.MethodPut,
		URL:    url,
		Body:   body,
	}
}

// DELETE creates a DELETE request helper
func DELETE(url string) HTTPRequest {
	return HTTPRequest{
		Method: fiber.MethodDelete,
		URL:    url,
	}
}

// PATCH creates a PATCH request helper
func PATCH(url string, body interface{}) HTTPRequest {
	return HTTPRequest{
		Method: fiber.MethodPatch,
		URL:    url,
		Body:   body,
	}
}

// WithHeaders adds headers to the request
func (r HTTPRequest) WithHeaders(headers map[string]string) HTTPRequest {
	r.Headers = headers
	return r
}

// WithContentType sets the content type
func (r HTTPRequest) WithContentType(contentType string) HTTPRequest {
	r.ContentType = contentType
	return r
}
