package model_test

import (
	"testing"

	. "webtyp.com/model"
)

func TestVerb_OfEachAction(t *testing.T) {
	for action, want := range map[Action]string{Create: "create", Read: "read", Update: "update", Delete: "delete"} {
		if got := action.Verb(); got != want {
			t.Fatalf("%v.Verb() = %q, want %q", action, got, want)
		}
	}
}

func TestVerb_OfASetOrNothingIsEmpty(t *testing.T) {
	for _, a := range []Action{0, Create | Read, AllActions} {
		if got := a.Verb(); got != "" {
			t.Fatalf("%v.Verb() = %q, want empty (a set is not one action)", a, got)
		}
	}
}

func TestParseVerb_RoundTrips(t *testing.T) {
	for _, a := range []Action{Create, Read, Update, Delete} {
		got, ok := ParseVerb(a.Verb())
		if !ok || got != a {
			t.Fatalf("ParseVerb(%q) = %v, %v; want %v, true", a.Verb(), got, ok, a)
		}
	}
}

func TestParseVerb_UnknownIsNotAnAction(t *testing.T) {
	for _, word := range []string{"", "patch", "Create", "c", "create "} {
		if got, ok := ParseVerb(word); ok || got != 0 {
			t.Fatalf("ParseVerb(%q) = %v, %v; want 0, false (exact lowercase word only)", word, got, ok)
		}
	}
}
