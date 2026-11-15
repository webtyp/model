package model

// When says under which values of another field a field is shown: the other
// field's name and the values that show this one. A field that carries a When
// (Field.ShownWhen) is hidden, emptied and not validated while the other field
// holds any other value.
//
// Example: the issuing country matters only for a passport or a foreign id:
//
//	{Name: "document_country", ShownWhen: model.When{Field: "document_type", In: []string{"passport", "foreign_id"}}}
type When struct {
	Field string   // name of the other field of the same definition
	In    []string // the values of that field for which this one is shown
}

// IsZero reports whether the field is always shown (no When declared).
func (w When) IsZero() bool { return w.Field == "" }

// Holds reports whether the other field's current value shows this field. A
// zero When always holds.
func (w When) Holds(value string) bool {
	if w.IsZero() {
		return true
	}
	for _, v := range w.In {
		if v == value {
			return true
		}
	}
	return false
}
