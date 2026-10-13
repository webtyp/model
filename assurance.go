package model

// Assurance is how strongly the identity of a request was proven — the
// authenticator assurance levels of NIST SP 800-63B, reduced to what an
// enforcer has to tell apart.
//
// The authentication layer sets it next to the user id; a route that demands
// a second factor compares against it. Levels are ordered, so an enforcer
// writes `a < AssuranceMultiFactor`, never a list of accepted values.
type Assurance uint8

const (
	// AssuranceNone (THE ZERO VALUE) is an anonymous request: it proves nothing.
	AssuranceNone Assurance = iota

	// AssuranceSingleFactor is one factor: a session cookie, a client assertion.
	AssuranceSingleFactor

	// AssuranceMultiFactor adds a second, phishing-resistant factor: a passkey.
	AssuranceMultiFactor
)

// String renders the level as a word for logs and introspection. An unknown
// value reads as "none": a level nobody defined proves nothing.
func (a Assurance) String() string {
	switch a {
	case AssuranceSingleFactor:
		return "single_factor"
	case AssuranceMultiFactor:
		return "multi_factor"
	default:
		return "none"
	}
}
