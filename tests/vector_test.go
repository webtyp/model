package model_test

import (
	"strings"
	"testing"

	. "webtyp.com/model"
)

func TestVector_StorageIsBlob(t *testing.T) {
	k := Vector(384)
	if k.Storage() != FieldBlob {
		t.Errorf("Storage() = %v, want %v", k.Storage(), FieldBlob)
	}
	if k.Name() != "vector" {
		t.Errorf("Name() = %q, want %q", k.Name(), "vector")
	}
}

func TestVector_Dim(t *testing.T) {
	k := Vector(384)
	d, ok := k.(Dimensional)
	if !ok {
		t.Fatal("Vector kind does not implement Dimensional")
	}
	if d.Dim() != 384 {
		t.Errorf("Dim() = %d, want 384", d.Dim())
	}
}

func TestVector_ValidateOK(t *testing.T) {
	f := Field{Name: "emb", Type: Vector(384)}
	if err := ValidateVector(f, make([]byte, 384*4)); err != nil {
		t.Errorf("ValidateVector(dim*4 bytes) = %v, want nil", err)
	}
}

func TestVector_ValidateWrongDim(t *testing.T) {
	f := Field{Name: "emb", Type: Vector(384)}
	for _, dim := range []int{383, 385} {
		err := ValidateVector(f, make([]byte, dim*4))
		if err == nil {
			t.Errorf("ValidateVector(%d dims) = nil, want error", dim)
			continue
		}
		if !strings.Contains(err.Error(), "emb") {
			t.Errorf("ValidateVector(%d dims) error %q does not mention field name", dim, err)
		}
	}
}

func TestVector_ValidateNotMultipleOfFour(t *testing.T) {
	f := Field{Name: "emb", Type: Vector(384)}
	if err := ValidateVector(f, make([]byte, 1537)); err == nil {
		t.Error("ValidateVector(1537 bytes) = nil, want error")
	}
}

func TestVector_ValidateEmpty(t *testing.T) {
	nullable := Field{Name: "emb", Type: Vector(384)}
	if err := ValidateVector(nullable, nil); err != nil {
		t.Errorf("ValidateVector(nil) nullable = %v, want nil", err)
	}
	if err := ValidateVector(nullable, []byte{}); err != nil {
		t.Errorf("ValidateVector(empty) nullable = %v, want nil", err)
	}
	required := Field{Name: "emb", Type: Vector(384), NotNull: true}
	if err := ValidateVector(required, nil); err == nil {
		t.Error("ValidateVector(nil) NotNull = nil, want error")
	}
	if err := ValidateVector(required, []byte{}); err == nil {
		t.Error("ValidateVector(empty) NotNull = nil, want error")
	}
}

func TestVector_ZeroPtr(t *testing.T) {
	f := Field{Name: "emb", Type: Vector(4)}
	var empty []byte
	if !IsZeroPtr(&empty, f.Type.Storage()) {
		t.Error("empty vector should be zero")
	}
	full := make([]byte, 4*4)
	if IsZeroPtr(&full, f.Type.Storage()) {
		t.Error("non-empty vector should not be zero")
	}
}

func TestBlob_StillNotDimensional(t *testing.T) {
	if _, ok := Blob().(Dimensional); ok {
		t.Error("Blob() must not satisfy Dimensional — the unconstrained escape hatch")
	}
}
