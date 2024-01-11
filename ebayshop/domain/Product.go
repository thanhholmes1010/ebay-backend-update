package domain

import (
	"ebayclone/changeset"
)

type Product struct {
	Id             uint32
	Name           string
	ProductTypeRel *ProductType // when have Rel keyword mean relation
}

func (p *Product) Validators() map[string]*changeset.Box {
	return map[string]*changeset.Box{
		"Id":             changeset.NewBox().Ops(changeset.AI),
		"Name":           changeset.NewBox().Ops(changeset.NotNullable),
		"ProductTypeRel": changeset.NewBox().Ops(changeset.NotNullable).SetEmbeddedClass(&ProductType{}, "Id"),
	}
}
