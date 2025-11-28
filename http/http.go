package http

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type (
	// Client = http.Client
	ResponseWriter = http.ResponseWriter
	Request        = http.Request
	Handler        = http.Handler
	HandlerFunc    = http.HandlerFunc
	Transport      = http.Transport
)

var NewRequest = http.NewRequest

// ServeFile serves files from the filesystem
func ServeFile(w ResponseWriter, r *Request, name string) {
	http.ServeFile(w, r, name)
}

// MaxBytesReader limits the size of the request body
func MaxBytesReader(w ResponseWriter, r io.ReadCloser, n int64) io.ReadCloser {
	return http.MaxBytesReader(w, r, n)
}

const (
	StatusOK                    = http.StatusOK
	StatusCreated               = http.StatusCreated
	StatusInternalServerError   = http.StatusInternalServerError
	StatusBadRequest            = http.StatusBadRequest
	StatusUnauthorized          = http.StatusUnauthorized
	StatusForbidden             = http.StatusForbidden
	StatusNotFound              = http.StatusNotFound
	StatusMethodNotAllowed      = http.StatusMethodNotAllowed
	StatusConflict              = http.StatusConflict
	StatusServiceUnavailable    = http.StatusServiceUnavailable
	StatusUnprocessableEntity   = http.StatusUnprocessableEntity
	StatusRequestEntityTooLarge = http.StatusRequestEntityTooLarge

	MethodGet     = http.MethodGet
	MethodPost    = http.MethodPost
	MethodPut     = http.MethodPut
	MethodDelete  = http.MethodDelete
	MethodPatch   = http.MethodPatch
	MethodHead    = http.MethodHead
	MethodOptions = http.MethodOptions

	ContentTypeJson     = "application/json"
	ContentTypeText     = "text/plain"
	ContentTypeHeader   = "Content-Type"
	AuthorizationHeader = "Authorization"
)

var client = &http.Client{
	Timeout: 300 * time.Second,
}

// createJsonRequest creates a new HTTP request with JSON body and headers
func createJsonRequest(method, url string, body any, headers map[string]string) (*http.Request, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	// Set content type for JSON request
	req.Header.Set(ContentTypeHeader, ContentTypeJson)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

// func doGet(url string, headers map[string]string) (*http.Response, error) {
// 	req, err := createJsonRequest("GET", url, nil, headers)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return client.Do(req)
// }

// doPost executes a POST request and returns the response
func doPost(url string, body any, headers map[string]string) (*http.Response, error) {
	req, err := createJsonRequest("POST", url, body, headers)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

func getFullUrl(path string) (string, error) {
	slog.Info("getFullUrl(): BEFORE", "path", path)
	u, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	slog.Info("getFullUrl(): AFTER", "url", u.String())
	return u.String(), nil
}

func Do(method, path string, headers map[string]string, body any) (data []byte, status int, err error) {
	var dataBytes []byte

	if headers == nil {
		headers = make(map[string]string)
	}
	switch body := body.(type) {
	case []byte:
		dataBytes = body
	default:
		dataBytes, err = json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		headers[ContentTypeHeader] = ContentTypeJson
	}

	fullUrl, err := getFullUrl(path)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest(method, fullUrl, bytes.NewReader(dataBytes))
	if err != nil {
		return nil, 0, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{
		Timeout: 300 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	return respBody, resp.StatusCode, nil
}

func Get(url string, headers map[string]string) (data []byte, status int, err error) {
	return Do(MethodGet, url, headers, nil)
}

func Post(url string, body []byte, headers map[string]string) (data []byte, status int, err error) {
	return Do(MethodPost, url, headers, body)
}

func Put(url string, body []byte, headers map[string]string) (data []byte, status int, err error) {
	return Do(MethodPut, url, headers, body)
}

func Patch(url string, body []byte, headers map[string]string) (data []byte, status int, err error) {
	return Do(MethodPatch, url, headers, body)
}

func Delete(url string, body []byte, headers map[string]string) (data []byte, status int, err error) {
	return Do(MethodDelete, url, headers, nil)
}

func GetJson(url string, headers map[string]string) (*http.Response, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := createJsonRequest("GET", url, nil, headers)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}

func GetJsonBytes(url string, headers map[string]string) ([]byte, int, error) {
	resp, err := GetJson(url, headers)
	if err != nil {
		return nil, 0, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	return body, resp.StatusCode, nil
}

func PostJson(url string, body any, headers map[string]string) (*http.Response, error) {
	return doPost(url, body, headers)
}

// PostJsonStreamResponse sends a POST request and returns a channel that streams JSON objects from the response
// The response is expected to contain newline-delimited JSON objects
func PostJsonStreamResponse(url string, body any, headers map[string]string) (<-chan []byte, error) {
	resp, err := doPost(url, body, headers)
	if err != nil {
		return nil, err
	}

	// Create channel for streaming JSON objects
	jsonChan := make(chan []byte, 100)

	go func() {
		defer resp.Body.Close()
		defer close(jsonChan)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			data := scanner.Bytes()
			if len(data) == 0 {
				continue
			}

			// Send raw bytes without unmarshaling
			jsonChan <- data
		}

		if err := scanner.Err(); err != nil {
			// Handle scanner error if needed
		}
	}()

	return jsonChan, nil
}

// PostJsonStreamResponseWithCallback sends a POST request and processes JSON objects from the response using a callback function
func PostJsonStreamResponseWithCallback(url string, body any, headers map[string]string, callback func(data []byte) error) error {
	resp, err := doPost(url, body, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		data := scanner.Bytes()
		if len(data) == 0 {
			continue
		}

		// Pass raw bytes to callback without unmarshaling
		if err := callback(data); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func WriteJson(w ResponseWriter, status int, body any) error {
	w.Header().Set(ContentTypeHeader, ContentTypeJson)
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(body)
}

func Error(w ResponseWriter, data any, status int) {
	var dataString string
	switch data := data.(type) {
	case string:
		dataString = data
	case error:
		dataString = data.Error()
	case map[string]any:
		dataBytes, err := json.Marshal(data)
		if err != nil {
			dataString = fmt.Sprintf("%s: %v", err.Error(), data)
		} else {
			dataString = string(dataBytes)
		}
	}
	w.Header().Set(ContentTypeHeader, ContentTypeJson)
	slog.Error("http.Error():", "data", data)
	http.Error(w, fmt.Sprintf(`{"error": "%s"}`, dataString), status)
}

func ErrorEx(w ResponseWriter, data any, err error, status int) {
	var dataString string
	switch data := data.(type) {
	case string:
		dataString = data
	case error:
		dataString = data.Error()
	case map[string]any:
		dataBytes, err := json.Marshal(data)
		if err != nil {
			dataString = fmt.Sprintf("%s: %v", err.Error(), data)
		} else {
			dataString = string(dataBytes)
		}
	}
	w.Header().Set(ContentTypeHeader, ContentTypeJson)
	slog.Error("http.Error():", "data", data)
	slog.Debug("http.Error():", "error", err)
	http.Error(w, fmt.Sprintf(`{"error": "%s"}`, dataString), status)
}
