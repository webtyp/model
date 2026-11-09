package model

import "webtyp.com/fmt"

// Fields is the ordered list of field definitions of a model.
// Alias (not a named type) so a []Field literal is assignable without conversion.
type Fields = []Field

// Definition is the hand-written, fully-typed source of truth for a model.
// From a Definition, the ormc generator produces the concrete Go struct and its
// zero-reflection plumbing (Schema/Pointers/EncodeFields/DecodeFields/List).
//
// This inverts the previous flow (struct + string tags → generated schema):
// the schema literal is now authored by hand and everything else is derived.
type Definition struct {
	Name   string // model identity: table name, ModelName(), route key
	Fields Fields // ordered schema; Kinds come from model (e.g. model.Text())
	// Wire marks a transport format (a JSON-RPC message, a webhook payload):
	// its field names are protocol, never screen text — langc does not
	// extract them and no form is built from it.
	Wire bool
	// Allow is what the table offers through the API (model.Create | model.Read | …). Zero = nothing: the table is
	// internal to its module. Read is not implied: a table only the module reads leaves it out.
	//
	//	var PostModel = model.Definition{Name: "post", Allow: model.Create | model.Read | model.Update, Fields: …}
	Allow Action
	// Rules says, for each action in Allow, whether the entity has a business rule of its own (AddMyBusinessRule) or crud's
	// checks are enough (HaveNoBusinessRule). An allowed action left undeclared is an error (definitioncheck), never a default.
	//
	//	Rules: model.Rules{Create: model.AddMyBusinessRule, Read: model.HaveNoBusinessRule, Update: model.AddMyBusinessRule}
	Rules Rules
	// MFA is the set of allowed actions that also require a second factor. It must be a subset of Allow (definitioncheck).
	//
	//	MFA: model.Update | model.Delete
	MFA Action
}

// Allower is implemented by the struct ormc generates from a Definition: what the table offers through the API.
// crud mounts only these actions and a screen shows only their buttons.
type Allower interface {
	Allow() Action
}

// TableSlice is a list of a table's rows that also tells what the table allows. ormc implements it on <Entity>List,
// so a screen asking for it does not compile with a list generated before Allow existed.
//
//	view.NewCallerLister(caller, ops, func() model.TableSlice { return &PatientList{} })
type TableSlice interface {
	ModelSlice
	Allower
}

// Field returns the field with the given name and true, or a zero Field and false.
func (d Definition) Field(name string) (Field, bool) {
	for _, f := range d.Fields {
		if f.Name == name {
			return f, true
		}
	}
	return Field{}, false
}

// Wirer is implemented by the struct ormc generates from a Definition with
// Wire: true. form.New refuses it: a transport format is never a form.
type Wirer interface {
	Wire() bool
}

// KeySeparator joins the values of a composite primary key into the one id a record has on the wire, in schema order.
// A key value never contains it (crud rejects it on create).
const KeySeparator = "/"

// JoinKey is the wire id of a composite key: JoinKey("r1", "e7") == "r1/e7".
func JoinKey(parts ...string) string {
	return fmt.JoinSlice(parts, KeySeparator)
}

// SplitKey is the inverse of JoinKey; ok is false when id does not have exactly n parts or a part is empty.
//
//	room, equipment, ok := …  // parts, ok := model.SplitKey("r1/e7", 2) → ["r1","e7"], true
func SplitKey(id string, n int) (parts []string, ok bool) {
	if n <= 0 || id == "" {
		return nil, false
	}
	parts = make([]string, 0, n)
	start := 0
	for i := 0; i < len(id); i++ {
		if id[i] == KeySeparator[0] {
			part := id[start:i]
			if part == "" {
				return nil, false
			}
			parts = append(parts, part)
			start = i + 1
		}
	}
	last := id[start:]
	if last == "" {
		return nil, false
	}
	parts = append(parts, last)
	if len(parts) != n {
		return nil, false
	}
	return parts, true
}
