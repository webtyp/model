package model

// RuleMode is what one allowed action does about the business. The zero value means "undeclared" and is refused by definitioncheck.
type RuleMode uint8

const (
	// AddMyBusinessRule: the entity has a business rule of its own, run before crud stores; definitionc creates its rule file to complete.
	AddMyBusinessRule RuleMode = iota + 1
	// HaveNoBusinessRule: no business rule of mine; crud validates the schema and stores, nothing to write.
	HaveNoBusinessRule
)

// Rules declares a RuleMode per action. Each field is for the action of the same name.
//
//	model.Rules{Create: model.AddMyBusinessRule, Read: model.HaveNoBusinessRule}
type Rules struct {
	Create RuleMode
	Read   RuleMode
	Update RuleMode
	Delete RuleMode
}

// Of returns the mode declared for exactly one action; a set of actions, or anything that is not one of the four, returns 0.
//
//	Rules{Create: AddMyBusinessRule}.Of(Create) // AddMyBusinessRule
func (r Rules) Of(a Action) RuleMode {
	switch a {
	case Create:
		return r.Create
	case Read:
		return r.Read
	case Update:
		return r.Update
	case Delete:
		return r.Delete
	}
	return 0
}
