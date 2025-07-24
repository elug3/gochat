package access

import "github.com/elug3/gochat/pkg/model"

type ContactsAccess struct{}

type Action string

const (
	ActionInvite       Action = "invite"
	ActionDeleteMember Action = "delete"
	ActionDeleteGroup  Action = "delete group"
)

const (
	RoleMember  model.Role = "member"
	RoleManager model.Role = "manager"
	RoleOwner   model.Role = "owner"
)

type Policy struct {
	act    model.Role
	tgt    model.Role
	action Action
}

var Policies = []Policy{
	{act: RoleOwner, action: ActionDeleteGroup},
	{act: RoleOwner, action: ActionInvite},
}

func (access *ContactsAccess) Can(act, tgt model.Role, action Action) bool {
	for _, policy := range Policies {
		if policy.act == act && policy.action == action {
			return true
		}
	}
	return false
}
