package models

import "slices"

type Scope string

// We use iota to define the scope levels, where ScopeAdmin has the highest level of access, followed by ScopeWrite and ScopeRead.
// It lets us easily compare the scope level is equal or lower than the required scope level for a given operation.
const (
	ScopeAdmin = iota
	ScopeWrite
	ScopeRead
)

func ParseScope(scope int) Scope {
	switch scope {
	case ScopeAdmin:
		return "admin"
	case ScopeWrite:
		return "write"
	case ScopeRead:
		return "read"
	default:
		return "unknown"
	}
}

func AuthorizeScope(scope int, allowedScopes ...int) bool {
	return slices.ContainsFunc(allowedScopes, func(s int) bool {
		if s <= scope {
			return true
		}
		return false
	})
}
