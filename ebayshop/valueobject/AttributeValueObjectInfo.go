package valueobject

// Response ....
type OptionValueRes struct {
	Id    OptionValueId `json:"id"`
	Value any           `json:"value"`
}

type OneAttributeObjectRes struct {
	Id           AttributeId       `json:"id"`
	Name         string            `json:"name"`
	OptionValues []*OptionValueRes `json:"option_values"`
}
type AttributesObjectRes struct {
	Attributes []*OneAttributeObjectRes `json:"attributes"`
}
