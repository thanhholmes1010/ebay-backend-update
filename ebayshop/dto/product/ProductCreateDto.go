package product

import (
	"ebayclone/valueobject"
)

type ProductCreateReq struct {
	ProductTypeId uint32                  `json:"product_type_id"`
	Fields        *valueobject.FieldsJSON `json:"fields"`
	Name          string                  `json:"name"`
}

type ProductCreateRes struct {
	Id uint32 `json:"id"`
}
