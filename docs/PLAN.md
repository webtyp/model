---
PLAN: "feat: typed refusals — an error says its kind (invalid, not found, conflict) and, when a field failed, which one"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 4035289667386009274
---

> Read AGENTS.md at the repo root before any change.

# PLAN — errors that say why, and which field

## Why

Every module invents its own `errors.go` (`type domainError string` + a `statusOf` switch) to turn
an error into an HTTP status, and they drifted: `device_manager` and `staff_manager` answer 400
with an **empty body**, `service_catalog` sends `err.Error()` even on a 500 (SQL text reaches the
browser). The screen cannot paint the field that failed because `ValidateFields` returns plain
text: the field name is only the first word of a sentence. Decision D22 of
`veltylabs/mjosefa-cms/docs/MVP_DEPLOYMENT_ROADMAP.md`: one piece answers errors for every
operation. This plan is step 1 of 3: the **vocabulary** lives here, in the zero-dependency
contract every layer imports. Step 2 (`webtyp/rpc.WriteError`) maps it to a status and a body;
step 3 deletes each module's `errors.go`.

## Design gate (api-design)

1. **Prior art.** Kubernetes `k8s.io/apimachinery/pkg/api/errors` (`NewNotFound`, `NewConflict`,
   `NewInvalid` with a `field.ErrorList`; the API server maps each kind to its status), gRPC status
   codes (`InvalidArgument`, `NotFound`, `AlreadyExists`, `Internal`) with
   `google.rpc.BadRequest.FieldViolation{field, description}`, and RFC 9457 Problem Details
   (`status`, `title`, `detail` + extension members for invalid params). All three: the **kind**
   is a closed set owned by the framework, the field is structured, the transport chooses the
   status. Here the same, with the kind in `model` because modules, `orm` and `rpc` all import it.
2. **Names** (read aloud): `model.NotFound("service not found")` — "a not-found refusal";
   `model.KindOf(err)` — "the kind of this error"; `model.FieldOf(err)` — "the field of this
   error". `ErrorInternal` is the zero value: an error nobody classified is internal, never
   shown to the user.
3. **Ledger.** Concepts +1 (`ErrorKind`) / −N (each module's `domainError` + `statusOf`, removed in
   step 3); ways to classify an error: from N to 1.
4. **Where.** `model`: zero dependencies, imported by modules, `orm`, `router`, `rpc`, `layout`.
5. **Deletes.** Nothing here; step 3 deletes every module's `domainError`, `ValidationError`,
   `statusOf` and `writeError`.

## Stage 1 — `error.go` (new, no build tag)

```go
package model

// ErrorKind says why an operation refused, so the transport can answer the right
// status without knowing the module. A closed set.
type ErrorKind uint8

const (
	// ErrorInternal (THE ZERO VALUE) is an error nobody classified: a bug or a broken
	// dependency. The caller sees a generic message; the detail goes to the log.
	ErrorInternal ErrorKind = iota
	// ErrorInvalid: the input is wrong; the caller can fix it (a field, a missing id).
	ErrorInvalid
	// ErrorNotFound: the thing asked for does not exist.
	ErrorNotFound
	// ErrorConflict: it cannot be done in the current state (duplicate, in use).
	ErrorConflict
)

// String renders the kind as a word: "internal", "invalid", "not found", "conflict".
func (k ErrorKind) String() string

// Kinded is an error that knows its kind.
type Kinded interface {
	error
	Kind() ErrorKind
}

// Invalid, NotFound and Conflict are refusals a module declares as constants:
//
//	const ErrNotFound model.NotFound = "service not found"
type Invalid string
type NotFound string
type Conflict string

func (e Invalid) Error() string      { return string(e) }
func (e Invalid) Kind() ErrorKind    { return ErrorInvalid }
func (e NotFound) Error() string     { return string(e) }
func (e NotFound) Kind() ErrorKind   { return ErrorNotFound }
func (e Conflict) Error() string     { return string(e) }
func (e Conflict) Kind() ErrorKind   { return ErrorConflict }

// FieldError is a validation failure of one field. Its text is the cause as before
// ("name required"); Field is the field's Name, so a form can paint it.
type FieldError struct {
	Field string
	Err   error
}

func (e *FieldError) Error() string   { return e.Err.Error() }
func (e *FieldError) Kind() ErrorKind { return ErrorInvalid }

// KindOf returns err's kind; an error that does not say is ErrorInternal.
//
//	switch model.KindOf(err) { case model.ErrorNotFound: … }
func KindOf(err error) ErrorKind {
	if k, ok := err.(Kinded); ok {
		return k.Kind()
	}
	return ErrorInternal
}

// FieldOf returns the field a validation error is about, or "" when it is not one.
func FieldOf(err error) string {
	if f, ok := err.(*FieldError); ok {
		return f.Field
	}
	return ""
}
```

`KindOf(nil)` returns `ErrorInternal`; callers only call it with a non-nil error. Type
assertions only — no `==` between interfaces, no reflection (wasm rule).

## Stage 2 — `field.go`, `permitted.go`: validation returns `*FieldError`

Every error that `Field.Validate` and `ValidateFields` return about a named field is wrapped as
`&FieldError{Field: <the field's Name>, Err: <the exact fmt.Err(...) built today>}`, so the
**text does not change**. Concretely:

- `Field.Validate` (`field.go:151`): each `return fmt.Err(...)` and each `return err` from
  `f.Type.Validate`, `validateLength`, `validateChars` becomes the wrapped form. Do it once:
  rename the current body to `func (f Field) validate(value string) error` and make
  `Validate` call it and wrap a non-nil result unless it is already a `*FieldError`.
- `ValidateFields` (`field.go:249`): the `fmt.Err(field.Name, …)` returns become wrapped; errors
  coming from `field.Validate(val)` are already wrapped; a nested struct's error is returned as
  is (it is already a `*FieldError` of the inner field).
- The vector checks (`field.go:212-220`) are about `f.Name` too: wrap them.
- `kind.go` errors are not wrapped there (they have no field name); they reach the caller
  through `Field.Validate`, which wraps them.

## Stage 3 — `tests/error_test.go` (new)

| Test | Does | Expected |
|---|---|---|
| `TestKindOf_Declared` | `KindOf(NotFound("x"))`, `Conflict`, `Invalid` | `ErrorNotFound`, `ErrorConflict`, `ErrorInvalid` |
| `TestKindOf_PlainIsInternal` | `KindOf(fmt.Err("boom"))` | `ErrorInternal` |
| `TestFieldOf_RequiredField` | `ValidateFields(ActionCreate, x)` on a test Fielder whose `name` field is `NotNull` and empty | `FieldOf(err) == "name"`, `KindOf(err) == ErrorInvalid`, `err.Error()` equals the text the existing tests expect for that case (copy it from `tests/field_test.go`) |
| `TestFieldOf_TooLong` | a field with `Permitted{Maximum: 3}`, value `"abcd"` | `FieldOf(err)` is that field's name |
| `TestFieldOf_NotAField` | `FieldOf(NotFound("x"))` | `""` |
| `TestErrorKind_String` | the four kinds | `"internal"`, `"invalid"`, `"not found"`, `"conflict"` |

Every existing test must stay green **without editing it**: the texts did not change.

## Stage 4 — docs

- `README.md`: a section `## Errors` with the four kinds, the module example
  (`const ErrNotFound model.NotFound = "service not found"`), `KindOf`, `FieldOf`, and "an
  unclassified error is internal: never shown to the user".
- `ARCHITECTURE.md`: why the kind lives in `model` (modules, `orm`, `rpc` import it; Kubernetes
  and gRPC precedent) and why the zero value is internal.
- If `lang.json` gains no key, nothing to translate (no new user-facing text).

## Delivery check

```
gotest
```

- A PR that does not compile, or whose tests could not run, is a failed delivery.
- Never pause to ask permission: when the check passes, commit, push and open the PR. Stuck 30
  minutes: open the PR listing what fails.
- `git diff --stat main` lists only: `error.go`, `field.go`, `permitted.go`, `tests/error_test.go`,
  `README.md`, `ARCHITECTURE.md`, this plan.

## Stages

| # | Files | Done when |
|---|---|---|
| 1 | `error.go` | compiles |
| 2 | `field.go`, `permitted.go` | existing tests green, unedited |
| 3 | `tests/error_test.go` | 6 tests pass |
| 4 | `README.md`, `ARCHITECTURE.md` | sections present |
