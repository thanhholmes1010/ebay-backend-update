package dto

import (
	"fmt"
	"net/http"
)

type BaseMessageResponse struct {
	StatusCode    int    `json:"status_code"`
	ErrCodeString string `json:"err_code_string"`
	ReponseObject any    `json:"reponse_object"`
}

func (b *BaseMessageResponse) TransformToStatusOk(responseObject any) {
	b.StatusCode = http.StatusOK
	b.ErrCodeString = ""
	b.ReponseObject = responseObject
}

func (b *BaseMessageResponse) TransformToNotFoundEntity(entityName string) {
	b.StatusCode = http.StatusNotFound
	b.ErrCodeString = fmt.Sprintf("Not Found Any Entity Type=[%v]", entityName)
	b.ReponseObject = nil
}
