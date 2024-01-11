package product_type_dto

import "ebayclone/valueobject"

// Request ....
type ProductTypeCreateReq struct {
	Name       string           `json:"name"`
	Attributes map[string][]any `json:"attributes"`
}

type ProductTypeCreateRes struct {
	Id         uint32                           `json:"id"`
	Attributes *valueobject.AttributesObjectRes `json:"attributesObjectRes"`
}
