package product_type_dto

import "ebayclone/domain"

type ProductTypeGetAllRes struct {
	ProductTypes []*domain.ProductType `json:"product_types"`
}
