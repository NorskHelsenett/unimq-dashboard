package models

import (
	"encoding/json"
	"testing"
)

func TestScopeAllows(t *testing.T) {
	if !ScopeAdmin.Allows(ScopeWrite) {
		t.Fatal("admin should allow write")
	}
	if ScopeRead.Allows(ScopeWrite) {
		t.Fatal("read should not allow write")
	}
}

func TestACLAllows(t *testing.T) {
	acl := ACL{
		Group:       "operators",
		Permissions: []Scope{ScopeWrite},
		VhostIDs:    []string{"orders"},
	}

	if !acl.Allows(ScopeRead, "orders") {
		t.Fatal("write permission should allow reading the configured vhost")
	}
	if acl.Allows(ScopeWrite, "billing") {
		t.Fatal("ACL should not allow an unconfigured vhost")
	}
}

func TestScopeJSON(t *testing.T) {
	var scope Scope
	if err := json.Unmarshal([]byte(`"write"`), &scope); err != nil {
		t.Fatalf("unmarshal scope: %v", err)
	}
	if scope != ScopeWrite {
		t.Fatalf("got %v, want write", scope)
	}

	data, err := json.Marshal(ScopeAdmin)
	if err != nil {
		t.Fatalf("marshal scope: %v", err)
	}
	if string(data) != `"admin"` {
		t.Fatalf("got %s, want admin", data)
	}
}
