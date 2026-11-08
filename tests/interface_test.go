package model_test

import "webtyp.com/model"

// modelStub mimics ormc's generated output: if this stub stops satisfying model.Model,
// the contract and the generator have drifted apart.
type modelStub struct{ Name string }

func (m *modelStub) ModelName() string                { return "model_stub" }
func (m *modelStub) Schema() []model.Field            { return nil }
func (m *modelStub) Pointers() []any                  { return []any{&m.Name} }
func (m *modelStub) IsNil() bool                      { return m == nil }
func (m *modelStub) EncodeFields(w model.FieldWriter) {}
func (m *modelStub) DecodeFields(r model.FieldReader) {}

var _ model.Model = (*modelStub)(nil)

// idGeneratorFunc adapts a plain function to model.IDGenerator — the shape a
// composition root's test double takes: no concrete generator, just the
// single method a consumer needs.
type idGeneratorFunc func() string

func (f idGeneratorFunc) NewID() string { return f() }

var _ model.IDGenerator = idGeneratorFunc(nil)

type tableSliceStub struct {
	modelStub
}

func (s *tableSliceStub) Len() int               { return 0 }
func (s *tableSliceStub) At(i int) model.Fielder { return nil }
func (s *tableSliceStub) Append() model.Fielder  { return nil }
func (s *tableSliceStub) Allow() model.Action    { return model.Read }

var _ model.TableSlice = (*tableSliceStub)(nil)

