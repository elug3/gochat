package access

import "slices"

type Action string

type Role string

const (
	ActionRead         Action = "read"
	ActionSend         Action = "send"
	ActionInvite       Action = "invite"
	ActionLeave        Action = "leave"
	ActionDeleteMember Action = "delete_member"
	ActionDeleteGroup  Action = "delete_group"
)

const (
	RoleMember  Role = "member"
	RoleManager Role = "manager"
	RoleOwner   Role = "owner"
)

type Policy struct {
	allow  bool
	target Role
}

var Policies = map[Role]map[Action]Policy{
	RoleMember: {
		ActionRead:  {allow: true},
		ActionSend:  {allow: true},
		ActionLeave: {allow: true},
	},
	RoleManager: {
		ActionInvite:       {allow: true},
		ActionDeleteMember: {target: RoleMember, allow: true},
	},
	RoleOwner: {
		ActionLeave:       {allow: false},
		ActionDeleteGroup: {allow: true},
		ActionDeleteMember: {
			allow:  true,
			target: RoleManager, // owners can remove managers and members, but not other owners
		},
	},
}

var RoleHierarchy = map[Role][]Role{
	RoleMember:  {},
	RoleManager: {RoleMember},
	RoleOwner:   {RoleManager},
}

func gaterRoles(role Role) []Role {
	res := make([]Role, 0)
	visited := make(map[Role]bool)
	var dfs func(role Role)

	dfs = func(role Role) {
		if visited[role] {
			return
		}
		visited[role] = true
		res = append(res, role)
		for _, node := range RoleHierarchy[role] {
			dfs(node)
		}
	}

	dfs(role)

	return res
}

func hierarchyPolicies(role Role) map[Action]Policy {
	res := make(map[Action]Policy)
	roles := gaterRoles(role)
	slices.Sort(roles)
	for _, r := range roles {
		if policy, exists := Policies[r]; exists {
			for action, p := range policy {
				res[action] = p
			}
		}
	}
	return res
}

func Can(act, tgt Role, action Action) bool {
	policies := hierarchyPolicies(act)
	p, ok := policies[action]
	if !ok {
		return false
	}
	if !p.allow {
		return false
	}
	if p.target == "" {
		return true
	}
	return canTarget(tgt, p.target)
}

// canTarget returns true when the target's role is at or below the allowed target role
// in the hierarchy (e.g. owners can target managers and members, managers can only
// target members).
func canTarget(target Role, max Role) bool {
	for _, allowed := range gaterRoles(max) {
		if allowed == target {
			return true
		}
	}
	return false
}
