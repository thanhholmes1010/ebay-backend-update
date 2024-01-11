package valueobject

type AttributeId uint32
type OptionValueId uint32
type FieldsJSON map[AttributeId]OptionValueId
type AggregateFieldJSON struct {
	Fields map[AttributeId]map[OptionValueId]int `json:"fields"`
}
