package model

import "webtyp.com/fmt"

// Item is ONE record shown as an entry of a list — the neutral form any renderer
// draws. ormc generates func (m *X) Item() Item from the fields marked Item:…;
// never write it by hand.
type Item struct {
	ID          string // the primary key, as text: the selection key
	Label       string // primary text (the field marked ItemLabel)
	Description string // secondary text (the fields marked ItemDescription, joined with JoinItem)
	Badge       string // leading badge (the field marked ItemBadge); a date/time field is seconds since 1970, as text
}

// Itemizer is a record that can be shown as a list entry. ormc implements it.
type Itemizer interface {
	Item() Item
}

// ItemSlot is the field of Item a model field fills. The zero value keeps it out.
type ItemSlot uint8

const (
	ItemNone        ItemSlot = iota // not shown in the list
	ItemLabel                       // Item.Label; exactly one per listed model
	ItemDescription                 // Item.Description; several are joined with JoinItem, in field order
	ItemBadge                       // Item.Badge; at most one
)

// itemSeparator joins several ItemDescription values.
const itemSeparator = " · "

// JoinItem joins the non-empty parts with " · ". Generated Item() methods call it
// when several fields are marked ItemDescription.
// Example: model.JoinItem("local", "", "Oficina") == "local · Oficina"
func JoinItem(parts ...string) string {
	var non []string
	for _, p := range parts {
		if p != "" {
			non = append(non, p)
		}
	}
	return fmt.JoinSlice(non, itemSeparator)
}

const (
	errItemTwoLabels  = "model: two fields marked ItemLabel: "
	errItemTwoBadges  = "model: two fields marked ItemBadge: "
	errItemNoLabel    = "model: fields marked ItemDescription or ItemBadge need a field marked ItemLabel"
	errItemNoPK       = "model: listed models need a primary key field"
	errRefAndOptions  = "model: a field cannot have both Ref and Options: "
)

// ValidateItem reports the first rule the fields break: at most one ItemLabel and
// one ItemBadge, an ItemLabel and a primary key whenever any field is marked, never
// Ref and Options on the same field. ormc calls it when it generates.
// Example: if err := model.ValidateItem(ServiceModel.Fields); err != nil { … }
func ValidateItem(fields []Field) error {
	label, badge, marked, pk := "", "", false, false
	for _, f := range fields {
		if f.Ref != nil && f.Options != nil {
			return fmt.Err(errRefAndOptions + f.Name)
		}
		if f.IsPK() {
			pk = true
		}
		switch f.Item {
		case ItemLabel:
			if label != "" {
				return fmt.Err(errItemTwoLabels + label + ", " + f.Name)
			}
			label = f.Name
			marked = true
		case ItemBadge:
			if badge != "" {
				return fmt.Err(errItemTwoBadges + badge + ", " + f.Name)
			}
			badge = f.Name
			marked = true
		case ItemDescription:
			marked = true
		}
	}
	if marked && label == "" {
		return fmt.Err(errItemNoLabel)
	}
	if marked && !pk {
		return fmt.Err(errItemNoPK)
	}
	return nil
}
