---
PLAN: "feat: Vector(dim) kind for embedding columns"
TAG: v0.2.0
EXECUTOR: unassigned
REVIEWER: none
---

> Part of the browser-native semantic search effort. Master index:
> https://github.com/webtyp/agent/blob/main/docs/PLAN.md — decision **D1** there is the
> rationale for everything below and is not re-argued here.

# Plan — a `Vector(dim)` kind

## Why

Semantic search needs a column that holds an embedding: a fixed-length `[]float32`
serialised little-endian. `model` can already carry the bytes — `FieldBlob` exists
(`field.go:13`) and is wired through `IsZeroPtr` (`field.go:310`) and `ValuesFrom`
(`field.go:383`) — but a plain `Blob()` column carries **no dimension**, so nothing
downstream can reject a 384-dim vector written into a 768-dim column. That mismatch is
silent, and it corrupts every subsequent search result rather than failing loudly.

This plan adds the dimension to the schema, where the rest of the ecosystem can read it.

## What does NOT change

**No new `FieldType`.** `FieldBlob` stays the storage type. Adding a
`FieldFloat32Slice` constant would force an edit to every exhaustive `switch` over
`FieldType` in this repository, in `storage/mem`, `sqlt`, `postgres` and `indexdb` —
a breaking change across six repositories that buys nothing `FieldBlob` does not already
provide. `Vector(dim)` is a `Kind`, and `Kind` is exactly the seam designed for
"same storage, different semantics" (it is what distinguishes `Text()` from an email
kind today).

The `Kind` interface is **not** modified. Its three methods stay as they are.

`ValuesFrom`, `IsZeroPtr`, `ScanFields` and the codecs need no edit: they dispatch on
`Field.Type.Storage()`, which for a vector returns `FieldBlob`, already handled.

## Changes

### 1. `kind.go` — the `Vector` constructor

Add below `Blob()`:

```go
// Dimensional is implemented by kinds whose value has a fixed element count.
// Consumers type-assert for it; a kind that does not implement it is unconstrained.
type Dimensional interface {
	Dim() int
}

type vectorKind struct {
	baseKind
	dim int
}

func (k vectorKind) Dim() int { return k.dim }

// Vector returns a kind for a fixed-length float32 embedding, stored as a
// little-endian blob of dim*4 bytes.
//
// Kind.Validate operates on a string and cannot see the bytes, so it always
// passes here; dimension enforcement is ValidateVector, called by the storage
// layer on the []byte itself.
func Vector(dim int) Kind {
	return vectorKind{
		baseKind: baseKind{
			storage: FieldBlob,
			name:    "vector",
			valid:   func(string) error { return nil },
		},
		dim: dim,
	}
}
```

`vectorKind` embeds `baseKind`, so it satisfies `Kind` with no extra method bodies.

### 2. `field.go` — byte-level validation

`Kind.Validate(value string) error` takes a string, which cannot express a blob
constraint. Rather than widen that interface (it is implemented by every kind and
called from `form`, `json` and `orm`), add a free function:

```go
// ValidateVector checks that b is a well-formed value for field f: a multiple of
// four bytes, and exactly f.Type.Dim()*4 bytes when the kind declares a dimension.
// A nil or empty b is accepted for a nullable field and rejected when f.NotNull.
func ValidateVector(f Field, b []byte) error {
	if len(b) == 0 {
		if f.NotNull {
			return fmt.Err("field", f.Name, "is required")
		}
		return nil
	}
	if len(b)%4 != 0 {
		return fmt.Err("field", f.Name, "vector length", len(b), "is not a multiple of 4")
	}
	if d, ok := f.Type.(Dimensional); ok && len(b)/4 != d.Dim() {
		return fmt.Err("field", f.Name, "expects", d.Dim(), "dimensions, got", len(b)/4)
	}
	return nil
}
```

Callers: `vectordb` before every write, and the `storage` conformance suite.

### 3. `field.go` — document the mapping

Extend the storage → Go type table in the `Field` doc comment (around line 72) with a
row for the vector kind:

```
// | FieldBlob (kind "vector") | []byte — dim*4 little-endian float32 |
```

### 4. `docs/` — no new document

The dimension contract is documented in the doc comments above. The README's type table
gains the same row.

## Tests

In `tests/field_test.go` and `tests/kind_permitted_override_test.go` style, standard
library only:

| Test | Asserts |
|---|---|
| `TestVector_StorageIsBlob` | `Vector(384).Storage() == FieldBlob` and `Name() == "vector"` |
| `TestVector_Dim` | the kind satisfies `Dimensional` and reports the constructed dim |
| `TestVector_ValidateOK` | `ValidateVector` accepts exactly `dim*4` bytes |
| `TestVector_ValidateWrongDim` | 383 and 385 dims are both rejected, with the field name in the message |
| `TestVector_ValidateNotMultipleOfFour` | a 1537-byte blob is rejected |
| `TestVector_ValidateEmpty` | empty passes when nullable, fails when `NotNull` |
| `TestVector_ZeroPtr` | `IsZeroPtr` on a `*[]byte` still behaves for a vector field |
| `TestBlob_StillNotDimensional` | `Blob()` does **not** satisfy `Dimensional` — the escape hatch survives |

## Acceptance checklist

```bash
grep -n "func Vector" kind.go            # → 1 match
grep -n "func ValidateVector" field.go   # → 1 match
grep -c "FieldFloat32\|FieldVector" *.go # → 0: no new FieldType was introduced
go vet ./...
gotest
```

Then release, because `storage` and `indexdb` both depend on this tag:

```bash
gopush 'feat: Vector(dim) kind for embedding columns'
```
