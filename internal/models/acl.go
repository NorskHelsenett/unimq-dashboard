package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
)

type Scope uint8

const (
	ScopeUnknown Scope = iota
	ScopeRead
	ScopeWrite
	ScopeAdmin
)

const AllVhosts = "*"

type ACL struct {
	Group       string   `json:"group" bson:"_id"`
	Permissions []Scope  `json:"permissions" bson:"permissions"`
	VhostIDs    []string `json:"vhost_ids" bson:"vhost_ids"`
}

func ParseScope(scope string) Scope {
	switch strings.ToLower(scope) {
	case "read":
		return ScopeRead
	case "write":
		return ScopeWrite
	case "admin":
		return ScopeAdmin
	default:
		return ScopeUnknown
	}
}

func (scope Scope) String() string {
	switch scope {
	case ScopeRead:
		return "read"
	case ScopeWrite:
		return "write"
	case ScopeAdmin:
		return "admin"
	default:
		return "unknown"
	}
}

func (scope Scope) Valid() bool {
	return scope >= ScopeRead && scope <= ScopeAdmin
}

func (scope Scope) Allows(required Scope) bool {
	return scope.Valid() && required.Valid() && scope >= required
}

func (scope Scope) MarshalJSON() ([]byte, error) {
	return json.Marshal(scope.String())
}

func (scope *Scope) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return errors.New("scope must be one of read, write, or admin")
	}

	parsed := ParseScope(value)
	if !parsed.Valid() {
		return fmt.Errorf("invalid scope %q", value)
	}
	*scope = parsed
	return nil
}

func (acl ACL) Validate() error {
	if strings.TrimSpace(acl.Group) == "" {
		return errors.New("group is required")
	}
	if len(acl.Permissions) == 0 {
		return errors.New("at least one permission is required")
	}
	for _, permission := range acl.Permissions {
		if !permission.Valid() {
			return fmt.Errorf("invalid permission %q", permission)
		}
	}
	if len(acl.VhostIDs) == 0 {
		return errors.New("at least one vhost_id is required")
	}
	return nil
}

func (acl ACL) Allows(required Scope, vhostID string) bool {
	if vhostID == "" || !slices.ContainsFunc(acl.Permissions, func(scope Scope) bool {
		return scope.Allows(required)
	}) {
		return false
	}

	return slices.Contains(acl.VhostIDs, vhostID) || slices.Contains(acl.VhostIDs, AllVhosts)
}
