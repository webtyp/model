package model_test

import (
	"testing"

	"webtyp.com/model"
)

// The instant kinds are stored as integers (so every backend and codec treats them like Int) and validate like Int.
func TestInstantKinds_StoreAsIntAndValidateLikeInt(t *testing.T) {
	for _, k := range []model.Kind{model.UnixTime(), model.NanoTime()} {
		if k.Storage() != model.FieldInt {
			t.Errorf("%s: storage %v, want FieldInt", k.Name(), k.Storage())
		}
		if err := k.Validate("1705332645"); err != nil {
			t.Errorf("%s: a number must validate: %v", k.Name(), err)
		}
		if err := k.Validate("12a"); err == nil {
			t.Errorf("%s: a non-number must be refused", k.Name())
		}
	}
	if model.UnixTime().Name() == model.NanoTime().Name() || model.UnixTime().Name() == model.Int().Name() {
		t.Error("each instant kind needs its own name: ormc picks the Go type by it")
	}
}
