package model_test

import (
	"testing"

	"webtyp.com/model"
)

func TestWhenZeroAlwaysHolds(t *testing.T) {
	var w model.When
	if !w.IsZero() || !w.Holds("anything") {
		t.Fatal("a zero When must always show its field")
	}
}

func TestWhenHoldsOnlyForListedValues(t *testing.T) {
	w := model.When{Field: "document_type", In: []string{"passport", "foreign_id"}}
	if !w.Holds("passport") || !w.Holds("foreign_id") {
		t.Error("a listed value must show the field")
	}
	if w.Holds("rut") || w.Holds("") {
		t.Error("an unlisted value must hide the field")
	}
}
