package utils

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Version   string      `json:"version"`
	ErrorCode int         `json:"errorCode"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
}

// Success 返回成功响应
func Success(w http.ResponseWriter, data interface{}) {
	ResponseJSON(w, http.StatusOK, "Success", data)
}

// Error 返回错误响应
func Error(w http.ResponseWriter, code int, message string, err error) {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	ResponseJSON(w, code, message, errMsg)
}

// BadRequest 返回400错误
func BadRequest(w http.ResponseWriter, message string) {
	ResponseJSON(w, http.StatusBadRequest, message, nil)
}

// InternalServerError 返回500错误
func InternalServerError(w http.ResponseWriter, err error) {
	ResponseJSON(w, http.StatusInternalServerError, "Internal Server Error", err.Error())
}

// ResponseJSON 返回JSON响应
func ResponseJSON(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	errCode := 0
	if code != http.StatusOK {
		errCode = code
	}
	json.NewEncoder(w).Encode(Response{
		Version:   "1",
		ErrorCode: errCode,
		Message:   message,
		Data:      data,
	})
}
