package utils

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
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
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}
