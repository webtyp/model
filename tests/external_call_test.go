package model_test

import (
	"testing"
	. "webtyp.com/model"
)

func TestExternalCallMode_ZeroValueIsUndeclared(t *testing.T) {
	var m ExternalCallMode
	if m != 0 {
		t.Fatalf("expected zero value to be 0, got %v", m)
	}
}

func TestExternalCall_WaitForReplyNeedsMaxWait(t *testing.T) {
	c1 := ExternalCallDefinition{Name: "charge", Mode: WaitForReply, MaxWaitSeconds: 10}
	if err := c1.Validate(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	c2 := ExternalCallDefinition{Name: "charge", Mode: WaitForReply, MaxWaitSeconds: 0}
	if err := c2.Validate(); err != ErrExternalCallMaxWaitRequired {
		t.Fatalf("expected ErrExternalCallMaxWaitRequired, got %v", err)
	}
}

func TestExternalCall_RetryInBackgroundNeedsMaxAttempts(t *testing.T) {
	c1 := ExternalCallDefinition{Name: "receipt", Mode: RetryInBackground, MaxAttempts: 8}
	if err := c1.Validate(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	c2 := ExternalCallDefinition{Name: "receipt", Mode: RetryInBackground, MaxAttempts: 0}
	if err := c2.Validate(); err != ErrExternalCallMaxAttemptsRequired {
		t.Fatalf("expected ErrExternalCallMaxAttemptsRequired, got %v", err)
	}
}

func TestExternalCall_WaitForConfirmationHasNoLimits(t *testing.T) {
	c1 := ExternalCallDefinition{Name: "settle", Mode: WaitForConfirmation}
	if err := c1.Validate(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	c2 := ExternalCallDefinition{Name: "settle", Mode: WaitForConfirmation, MaxWaitSeconds: 5}
	if err := c2.Validate(); err != ErrExternalCallLimitNotAllowed {
		t.Fatalf("expected ErrExternalCallLimitNotAllowed, got %v", err)
	}

	c3 := ExternalCallDefinition{Name: "settle", Mode: WaitForConfirmation, MaxAttempts: 3}
	if err := c3.Validate(); err != ErrExternalCallLimitNotAllowed {
		t.Fatalf("expected ErrExternalCallLimitNotAllowed, got %v", err)
	}
}

func TestExternalCall_LimitOfOtherModeRejected(t *testing.T) {
	c1 := ExternalCallDefinition{Name: "charge", Mode: WaitForReply, MaxWaitSeconds: 10, MaxAttempts: 3}
	if err := c1.Validate(); err != ErrExternalCallLimitNotAllowed {
		t.Fatalf("expected ErrExternalCallLimitNotAllowed, got %v", err)
	}

	c2 := ExternalCallDefinition{Name: "receipt", Mode: RetryInBackground, MaxAttempts: 8, MaxWaitSeconds: 10}
	if err := c2.Validate(); err != ErrExternalCallLimitNotAllowed {
		t.Fatalf("expected ErrExternalCallLimitNotAllowed, got %v", err)
	}
}

func TestExternalCall_NameRequired(t *testing.T) {
	c := ExternalCallDefinition{Mode: WaitForConfirmation}
	if err := c.Validate(); err != ErrExternalCallNameRequired {
		t.Fatalf("expected ErrExternalCallNameRequired, got %v", err)
	}
}

func TestExternalCall_ModeUndeclared(t *testing.T) {
	c1 := ExternalCallDefinition{Name: "charge"}
	if err := c1.Validate(); err != ErrExternalCallModeUndeclared {
		t.Fatalf("expected ErrExternalCallModeUndeclared, got %v", err)
	}

	c2 := ExternalCallDefinition{Name: "charge", Mode: 9}
	if err := c2.Validate(); err != ErrExternalCallModeUndeclared {
		t.Fatalf("expected ErrExternalCallModeUndeclared, got %v", err)
	}
}

func TestDefinition_ValidateExternalCalls(t *testing.T) {
	d1 := Definition{
		Name: "payment",
		ExternalCalls: []ExternalCallDefinition{
			{Name: "charge", Mode: WaitForReply, MaxWaitSeconds: 10},
			{Name: "receipt", Mode: RetryInBackground, MaxAttempts: 8},
		},
	}
	if err := d1.ValidateExternalCalls(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	d2 := Definition{Name: "payment"}
	if err := d2.ValidateExternalCalls(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestDefinition_DuplicateExternalCallName(t *testing.T) {
	d := Definition{
		Name: "payment",
		ExternalCalls: []ExternalCallDefinition{
			{Name: "charge", Mode: WaitForReply, MaxWaitSeconds: 10},
			{Name: "charge", Mode: WaitForConfirmation},
		},
	}
	if err := d.ValidateExternalCalls(); err != ErrExternalCallDuplicate {
		t.Fatalf("expected ErrExternalCallDuplicate, got %v", err)
	}
}

func TestDefinition_ExternalCallErrorsAreInvalid(t *testing.T) {
	errors := []error{
		ErrExternalCallNameRequired,
		ErrExternalCallModeUndeclared,
		ErrExternalCallMaxWaitRequired,
		ErrExternalCallMaxAttemptsRequired,
		ErrExternalCallLimitNotAllowed,
		ErrExternalCallDuplicate,
	}

	for _, err := range errors {
		invalid, ok := err.(Invalid)
		if !ok || invalid.Kind() != ErrorInvalid {
			t.Fatalf("expected error %v to be of Kind ErrorInvalid, got %v", err, invalid.Kind())
		}
	}
}
