package types

import (
	"github.com/nyaruka/goflow/assets"
)

// Field is a JSON serializable implementation of a field asset
type Field struct {
	Key_  string           `json:"key" validate:"required"`
	Name_ string           `json:"name"`
	Type_ assets.FieldType `json:"value_type" validate:"required"`
}

// NewField creates a new field from the passed in key, name and type
func NewField(key string, name string, valueType assets.FieldType) assets.Field {
	return &Field{Key_: key, Name_: name, Type_: valueType}
}

// Key returns the unique key of the field
func (f *Field) Key() string { return f.Key_ }

// Name returns the name of the field
func (f *Field) Name() string { return f.Name_ }

// Type returns the value type of the field
func (f *Field) Type() assets.FieldType { return f.Type_ }