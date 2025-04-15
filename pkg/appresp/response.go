package appresp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var ErrNotFound = AppError{"not found", 404}
var ErrServerInternal = AppError{"server internal error", 500}

type AppError struct {
	Message string
	Code    int
}

func (err *AppError) Error() string {
	return err.Message
}

type AppResponse struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func writeResponse(w http.ResponseWriter, data any, err error) {
	if err != nil {
		var appErr *AppError
		if errors.As(err, &appErr) {
			// response appErr
			ResponseFailed(w, appErr)
			return
		}
		// not handled error
		fmt.Println(err)
		ResponseFailed(w, &ErrServerInternal)
		return
	}
	ResponseSuccess(w, data)
}

func ResponseSuccess(w http.ResponseWriter, data any) {
	w.WriteHeader(200)
	encoder := json.NewEncoder(w)
	encoder.Encode(AppResponse{
		Success: true,
		Code:    200,
		Message: "",
		Data:    data,
	})
}

func ResponseFailed(w http.ResponseWriter, err *AppError) {
	w.WriteHeader(err.Code)
	encoder := json.NewEncoder(w)
	encoder.Encode(AppResponse{
		Success: false,
		Code:    err.Code,
		Message: err.Error(),
		Data:    nil,
	})
}

func HandleNotFound(w http.ResponseWriter, r *http.Request) {
	ResponseFailed(w, &ErrNotFound)
}
