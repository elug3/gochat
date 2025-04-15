package appresp

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func decodeResponse(t *testing.T, body []byte) AppResponse {
	var resp AppResponse
	err := json.Unmarshal(body, &resp)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return resp
}

func TestResponseSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]string{"key": "value"}
	ResponseSuccess(rec, data)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed reading body: %v", err)
	}

	resp := decodeResponse(t, body)
	if !resp.Success {
		t.Errorf("expected success true")
	}
	if resp.Code != http.StatusOK {
		t.Errorf("expected code %d, got %d", http.StatusOK, resp.Code)
	}
	if resp.Data == nil {
		t.Errorf("expected non-nil data")
	}
}

func TestResponseFailed(t *testing.T) {
	rec := httptest.NewRecorder()
	errApp := &AppError{Message: "bad request", Code: http.StatusBadRequest}
	ResponseFailed(rec, errApp)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed reading body: %v", err)
	}

	resp := decodeResponse(t, body)
	if resp.Success {
		t.Errorf("expected success false")
	}
	if resp.Code != http.StatusBadRequest {
		t.Errorf("expected code %d, got %d", http.StatusBadRequest, resp.Code)
	}
	if resp.Message != "bad request" {
		t.Errorf("expected message %q, got %q", "bad request", resp.Message)
	}
}

func TestWriteResponse_Success(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]int{"id": 123}

	WriteResponse(rec, data, nil)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed reading body: %v", err)
	}

	resp := decodeResponse(t, body)
	if !resp.Success {
		t.Errorf("expected success true")
	}
	if resp.Data == nil {
		t.Errorf("expected data set")
	}
}

func TestWriteResponse_WithAppError(t *testing.T) {
	rec := httptest.NewRecorder()
	errApp := &AppError{Message: "forbidden", Code: http.StatusForbidden}

	WriteResponse(rec, nil, errApp)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed reading body: %v", err)
	}

	resp := decodeResponse(t, body)
	if resp.Success {
		t.Errorf("expected success false")
	}
	if resp.Message != "forbidden" {
		t.Errorf("expected message %q, got %q", "forbidden", resp.Message)
	}
}

func TestWriteResponse_WithGenericError(t *testing.T) {
	rec := httptest.NewRecorder()
	genErr := errors.New("some error")

	WriteResponse(rec, nil, genErr)

	res := rec.Result()
	defer res.Body.Close()

	// Should respond with ErrServerInternal error
	if res.StatusCode != ErrServerInternal.Code {
		t.Errorf("expected status %d, got %d", ErrServerInternal.Code, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed reading body: %v", err)
	}

	resp := decodeResponse(t, body)
	if resp.Success {
		t.Errorf("expected success false")
	}
	if resp.Message != ErrServerInternal.Message {
		t.Errorf("expected message %q, got %q", ErrServerInternal.Message, resp.Message)
	}
}

func TestHandleNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	HandleNotFound(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed reading body: %v", err)
	}

	resp := decodeResponse(t, body)
	if resp.Success {
		t.Errorf("expected success false")
	}
	if resp.Code != http.StatusNotFound {
		t.Errorf("expected code %d, got %d", http.StatusNotFound, resp.Code)
	}
	if resp.Message != ErrNotFound.Message {
		t.Errorf("expected message %q, got %q", ErrNotFound.Message, resp.Message)
	}
}
