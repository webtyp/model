package model

import "webtyp.com/fmt"

// FieldTextSearch is the hidden column where the framework keeps each row's search text: every field that
// enters it (TextSearchFields), in its search form. No module declares it; a field with this name is refused
// at generation (definitioncheck).
const FieldTextSearch = "text_search"

// textSearchSeparator joins two fields so that their words never merge: «·» ends a field for fmt ToSearchText.
const textSearchSeparator = " · "

// SearchableText is implemented by a kind whose value is short free text a person types to find a record:
// a name, a phone, an email, a document number (webtyp.com/input: Text, Phone, Email, Rut). Options, numbers,
// dates, files, passwords and long texts (input.Textarea) do not implement it.
//
//	func (t *text) SearchableText() {}
type SearchableText interface {
	Kind
	SearchableText()
}

// TextSearchFields returns, in schema order, the indexes of the fields whose value enters the row's
// text_search: a SearchableText kind that is neither a primary key nor ServerOnly. Empty → the table has
// no text_search column.
//
//	idx := model.TextSearchFields(p.Schema()) // rut, name, phone, email of a patient
func TextSearchFields(schema []Field) []int {
	var idx []int
	for i, f := range schema {
		if f.IsPK() || f.ServerOnly {
			continue
		}
		if _, ok := f.Type.(SearchableText); ok {
			idx = append(idx, i)
		}
	}
	return idx
}

// TextSearchOf is the text_search value of one row: the values of its TextSearchFields, joined so that two
// fields never merge, in their search form. ptrs only needs a *string at those indexes.
//
//	model.TextSearchOf(p.Schema(), p.Pointers()) // "lucia perez 130479800 56912345678"
func TextSearchOf(schema []Field, ptrs []any) string {
	var parts []string
	for _, i := range TextSearchFields(schema) {
		if p, ok := ptrs[i].(*string); ok && p != nil && *p != "" {
			parts = append(parts, *p)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return fmt.Convert(fmt.JoinSlice(parts, textSearchSeparator)).ToSearchText().String()
}

// SearchFirst is implemented by the struct definitionc generates from a Definition with
// ListsOnlyAfterSearch: a read without text returns no rows (webtyp.com/crud) and the screen says
// «Type to search» (webtyp.com/layout/crudview).
//
//	if sf, ok := rec.(model.SearchFirst); ok && sf.ListsOnlyAfterSearch() { … }
type SearchFirst interface {
	ListsOnlyAfterSearch() bool
}
