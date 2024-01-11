package domain

import (
	"ebayclone/changeset"
	"ebayclone/valueobject"
)

type ProductType struct {
	Id              uint32
	Name            string
	Attributes      *valueobject.AttributesObjectRes
	AggregateFields *valueobject.AggregateFieldJSON
}

func (p *ProductType) Validators() map[string]*changeset.Box {
	return map[string]*changeset.Box{
		"Id":              changeset.NewBox().Ops(changeset.AI),
		"Name":            changeset.NewBox().Ops(changeset.NotNullable),
		"Attributes":      changeset.NewBox().Ops(changeset.NotNullable).JSONField(),
		"AggregateFields": changeset.NewBox().Ops(changeset.NotNullable).JSONField(),
	}
}

func (p *ProductType) CloneProductType() *ProductType {
	return &ProductType{
		Id:              p.Id,
		Name:            p.Name,
		Attributes:      p.Attributes,
		AggregateFields: p.AggregateFields,
	}
}
