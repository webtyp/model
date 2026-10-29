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
func (k ErrorKind) String() string {
	switch k {
	case ErrorInvalid:
		return "invalid"
	case ErrorNotFound:
		return "not found"
	case ErrorConflict:
		return "conflict"
	default:
		return "internal"
	}
}

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

func (e Invalid) Error() string    { return string(e) }
func (e Invalid) Kind() ErrorKind  { return ErrorInvalid }
func (e NotFound) Error() string   { return string(e) }
func (e NotFound) Kind() ErrorKind { return ErrorNotFound }
func (e Conflict) Error() string   { return string(e) }
func (e Conflict) Kind() ErrorKind { return ErrorConflict }

// FieldError is a validation failure of one field. Its text is the cause as before
// ("name required"); Field is the field's Name, so a form can paint it.
type FieldError struct {
	Field string
	Err   error
}

func (e *FieldError) Error() string       { return e.Err.Error() }
func (e *FieldError) Kind() ErrorKind     { return ErrorInvalid }
func (e *FieldError) FailedField() string { return e.Field }

// Fielded is an error that knows which field failed: a FieldError on the server,
// the rpc client's error in the browser.
type Fielded interface {
	error
	FailedField() string
}

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
// The same call works on the server and in the browser (the rpc client's error is
// Fielded too).
func FieldOf(err error) string {
	if f, ok := err.(Fielded); ok {
		return f.FailedField()
	}
	return ""
}
