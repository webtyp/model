package model

// ExternalCallMode says what a rule waits for when it asks for an external call. The zero value
// means "undeclared" and Validate refuses it.
type ExternalCallMode uint8

const (
	// WaitForReply: the request waits for the provider's answer, up to MaxWaitSeconds. If the time
	// runs out the call stays queued and the background runner finishes it.
	WaitForReply ExternalCallMode = iota + 1
	// RetryInBackground: the request answers at once and the call is retried in the background, at most MaxAttempts times.
	RetryInBackground
	// WaitForConfirmation: the call stays queued until the provider confirms it; there is no time limit here.
	WaitForConfirmation
)

// ExternalCallDefinition declares one call to another system that a rule of the entity may ask for.
// Limits are written out, never defaulted; a limit that does not belong to the Mode is an error.
//
//	ExternalCalls: []model.ExternalCallDefinition{
//		{Name: "charge", Mode: model.WaitForReply, MaxWaitSeconds: 10},
//		{Name: "receipt", Mode: model.RetryInBackground, MaxAttempts: 8},
//		{Name: "settle", Mode: model.WaitForConfirmation},
//	}
type ExternalCallDefinition struct {
	Name           string // the call's own name, written once here: "charge"
	Mode           ExternalCallMode
	MaxWaitSeconds int64 // WaitForReply only, greater than zero
	MaxAttempts    int64 // RetryInBackground only, greater than zero
}

const (
	ErrExternalCallNameRequired        Invalid = "external call name required"
	ErrExternalCallModeUndeclared      Invalid = "external call mode undeclared: use WaitForReply, RetryInBackground or WaitForConfirmation"
	ErrExternalCallMaxWaitRequired     Invalid = "external call MaxWaitSeconds required: WaitForReply needs a maximum wait greater than zero"
	ErrExternalCallMaxAttemptsRequired Invalid = "external call MaxAttempts required: RetryInBackground needs a number of attempts greater than zero"
	ErrExternalCallLimitNotAllowed     Invalid = "external call limit not allowed: WaitForReply takes only MaxWaitSeconds, RetryInBackground only MaxAttempts, WaitForConfirmation none"
	ErrExternalCallDuplicate           Invalid = "external call name repeated in the same entity"
)

// Validate reports the first thing wrong with the declaration, in this order: name, mode, then the limits of the mode.
//
//	err := model.ExternalCallDefinition{Name: "charge", Mode: model.WaitForReply}.Validate() // ErrExternalCallMaxWaitRequired
func (c ExternalCallDefinition) Validate() error {
	if c.Name == "" {
		return ErrExternalCallNameRequired
	}
	switch c.Mode {
	case WaitForReply:
		if c.MaxAttempts != 0 {
			return ErrExternalCallLimitNotAllowed
		}
		if c.MaxWaitSeconds <= 0 {
			return ErrExternalCallMaxWaitRequired
		}
	case RetryInBackground:
		if c.MaxWaitSeconds != 0 {
			return ErrExternalCallLimitNotAllowed
		}
		if c.MaxAttempts <= 0 {
			return ErrExternalCallMaxAttemptsRequired
		}
	case WaitForConfirmation:
		if c.MaxWaitSeconds != 0 || c.MaxAttempts != 0 {
			return ErrExternalCallLimitNotAllowed
		}
	default:
		return ErrExternalCallModeUndeclared
	}
	return nil
}

// ValidateExternalCalls validates every declared call and refuses two with the same name.
func (d Definition) ValidateExternalCalls() error {
	for i := range d.ExternalCalls {
		if err := d.ExternalCalls[i].Validate(); err != nil {
			return err
		}
		for j := 0; j < i; j++ {
			if d.ExternalCalls[j].Name == d.ExternalCalls[i].Name {
				return ErrExternalCallDuplicate
			}
		}
	}
	return nil
}
