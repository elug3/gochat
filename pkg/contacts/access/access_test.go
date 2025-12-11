package access

import (
	"slices"
	"testing"
)

func TestCan(t *testing.T) {
	testCases := []struct {
		role   Role
		tgt    Role
		action Action
		want   bool
	}{
		{RoleMember, RoleMember, ActionRead, true},
		{RoleOwner, RoleOwner, ActionLeave, false},
		{RoleOwner, RoleOwner, ActionDeleteGroup, true},
		{RoleManager, RoleOwner, ActionDeleteMember, false},
		{RoleOwner, RoleManager, ActionDeleteMember, true},
		{RoleOwner, RoleOwner, ActionDeleteMember, false},
	}

	for _, tc := range testCases {
		got := Can(tc.role, tc.tgt, tc.action)
		if got != tc.want {
			t.Errorf("For role %v and target %v with action %v: expected %v but got %v", tc.role, tc.tgt, tc.action, tc.want, got)
		}
	}
}

func TestGaterRoles(t *testing.T) {
	testCases := []struct {
		role Role
		want []Role
	}{
		{RoleMember, []Role{RoleMember}},
		{RoleManager, []Role{RoleManager, RoleMember}},
		{RoleOwner, []Role{RoleOwner, RoleManager, RoleMember}},
	}

	for _, tc := range testCases {
		got := gaterRoles(tc.role)
		if len(tc.want) != len(got) {
			t.Errorf("For role %v, expected %v but got %v", tc.role, tc.want, got)
		}
	}
}

func TestHierarchyPolicies(t *testing.T) {
	testCases := []struct {
		role Role
		want []Action
	}{
		{RoleMember, []Action{ActionRead, ActionSend, ActionLeave}},
		{RoleManager, []Action{ActionRead, ActionSend, ActionLeave, ActionInvite, ActionDeleteMember}},
		{RoleOwner, []Action{ActionRead, ActionSend, ActionLeave, ActionInvite, ActionDeleteMember, ActionDeleteGroup}},
	}

	for _, tc := range testCases {
		got := hierarchyPolicies(tc.role)
		want := slices.Compact(tc.want)

		if len(want) != len(got) {
			t.Errorf("For role %v, expected %v but got %v", tc.role, want, got)
		} else {
			for _, action := range want {
				if _, exists := got[action]; !exists {
					t.Errorf("For role %v, expected action %v but it was not found", tc.role, action)
				}
			}
		}
	}
}
