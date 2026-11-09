package model_test

import (
	"testing"

	. "webtyp.com/model"
)

func TestField_IsRevision(t *testing.T) {
	revision := Field{Name: "revision", Type: Int(), DB: &FieldDB{Revision: true}}
	if !revision.IsRevision() {
		t.Error("a field with DB.Revision must be the revision")
	}
	for name, f := range map[string]Field{
		"no DB":    {Name: "n", Type: Int()},
		"primary":  {Name: "id", Type: Text(), DB: &FieldDB{PK: true}},
		"unique":   {Name: "code", Type: Text(), DB: &FieldDB{Unique: true}},
		"empty DB": {Name: "n", Type: Int(), DB: &FieldDB{}},
	} {
		if f.IsRevision() {
			t.Errorf("%s: must not be the revision", name)
		}
	}
}
