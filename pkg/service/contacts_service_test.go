package service

import (
	"errors"
	"fmt"
	"testing"

	"github.com/elug3/gochat/internal/config"
	"github.com/elug3/gochat/pkg/model"
	"github.com/elug3/gochat/pkg/store"
	"github.com/elug3/gochat/pkg/store/contacts/sqlite"
)

func NewTestContactsService() (*ContactsService, error) {
	store, err := sqlite.NewContactsStore(&config.Config{
		NoSave: true,
	})
	if err != nil {
		return nil, err
	}
	// access, err := access.NewContactsAccess(store)
	// if err != nil {
	// 	return nil, err
	// }
	s, err := NewContactsService(store)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func TestContacts_CreateProfile(t *testing.T) {
	type Row struct {
		id      int
		name    string
		wantErr error
	}
	testCases := map[string]struct {
		rows []Row
	}{
		"create profile": {
			rows: []Row{
				{id: 1, name: "test", wantErr: nil},
			},
		},
		"create with same id": {
			rows: []Row{
				{id: 1, name: "a", wantErr: nil},
				{id: 1, name: "b", wantErr: store.ErrExists},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			s, _, err := setup(t, nil)
			if err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			for i, row := range tc.rows {
				p, err := s.CreateProfile(row.id, row.name)
				if !errors.Is(err, row.wantErr) {
					t.Errorf("row_%d: expected error: %q, got: %q", i, row.wantErr, err)
					continue
				}
				if err == nil && p.Id != row.id {
					t.Errorf("row_%d: expected id: %d, got: %d", i, row.id, p.Id)
				}
			}
		})
	}
}

func TestContacts_DeleteProfile(t *testing.T) {
	type DeleteProfile struct {
		profile string
		wantErr error
	}
	testCases := map[string]struct {
		preset *Preset
		rows   []DeleteProfile
	}{
		"delete profile": {
			preset: &Preset{
				profiles: map[string]presetProfile{
					"a": {userId: 1, name: "test"},
				},
			},
			rows: []DeleteProfile{
				{profile: "a", wantErr: nil},
			},
		},
		"delete non-existent profile": {
			preset: &Preset{
				profiles: map[string]presetProfile{
					"a": {userId: 1, name: "test"},
				},
			},
			rows: []DeleteProfile{
				{profile: "a", wantErr: nil},
			},
		},
		// TODO
		"delete profile with leave groups": {
			preset: &Preset{
				profiles: map[string]presetProfile{
					"a": {userId: 1, name: "test"},
				},
				groups: map[string]presetGroup{
					"g1": {name: "my group", owner: "a"},
					"g2": {name: "guest group", owner: "a"},
				},
			},
			rows: []DeleteProfile{
				{profile: "a", wantErr: nil},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			s, result, err := setup(t, tc.preset)
			if err != nil {
				t.Fatalf("setup failed: %v", err)
			}
			for i, row := range tc.rows {
				p, err := result.GetProfile(row.profile)
				if err != nil {
					t.Fatalf("row_%d: result.GetProfile: %q", i, err)
				}
				if err = s.DeleteProfile(p.Id); !errors.Is(err, row.wantErr) {
					t.Fatalf("expected error: %q, but got: %q", row.wantErr, err)
				}
			}
		})
	}
}

func TestContacts_Invite(t *testing.T) {
	type Invite struct {
		group   string
		inviter string
		invitee string
		wantErr error
	}
	preset := &Preset{
		profiles: map[string]presetProfile{
			"p1": {userId: 1, name: "p1"},
			"p2": {userId: 2, name: "p2"},
			"p3": {userId: 3, name: "p3"},
		},
		groups: map[string]presetGroup{
			"g1": {name: "test group", owner: "p1", member: []string{"p2"}},
		},
	}
	testCasess := map[string]struct {
		preset *Preset
		rows   []Invite
	}{
		"owner invite": {
			preset: preset,
			rows: []Invite{
				{group: "g1", inviter: "p1", invitee: "p3"},
			},
		},
		"member cannot invite": {
			preset: preset,
			rows: []Invite{
				{group: "g1", inviter: "p2", invitee: "p3", wantErr: store.ErrPermissionDenied},
			},
		},
		"already exists": {
			preset: preset,
			rows: []Invite{
				{group: "g1", inviter: "p1", invitee: "p3", wantErr: nil},
				{group: "g1", inviter: "p1", invitee: "p3", wantErr: store.ErrExists},
			},
		},
	}

	for name, tc := range testCasess {
		t.Run(name, func(t *testing.T) {
			s, result, err := setup(t, tc.preset)
			if err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			for i, row := range tc.rows {
				g, err := result.GetGroup(row.group)
				if err != nil {
					t.Fatal(err)
				}
				inviter, err := result.GetProfile(row.inviter)
				if err != nil {
					t.Fatal(err)
				}
				invitee, err := result.GetProfile(row.invitee)
				if err != nil {
					t.Fatal(err)
				}
				_, err = s.Invite(g.Id, inviter.Id, invitee.Id)
				if !errors.Is(err, row.wantErr) {
					t.Errorf("row_%d: expected error %q, but got %q", i, row.wantErr, err)
				}
			}
		})
	}
}

func TestContacts_DeleteGroup(t *testing.T) {
	type DeleteGroup struct {
		group   string
		profile string
		wantErr error
	}
	testCases := map[string]struct {
		preset *Preset
		rows   []DeleteGroup
	}{
		"owner can delete group": {
			preset: &Preset{
				profiles: map[string]presetProfile{
					"p1": {userId: 1, name: "p1"},
				},
				groups: map[string]presetGroup{
					"g1": {name: "test group", owner: "p1"},
				},
			},
			rows: []DeleteGroup{
				{group: "g1", profile: "p1"},
			},
		},
		"member cannot delete group": {
			preset: &Preset{
				profiles: map[string]presetProfile{
					"p1": {userId: 1, name: "p1"},
					"p2": {userId: 2, name: "test member"},
				},
				groups: map[string]presetGroup{
					"g1": {name: "test group", owner: "p1", member: []string{"p2"}},
				},
			},
			rows: []DeleteGroup{
				{group: "g1", profile: "p2", wantErr: store.ErrPermissionDenied},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			s, result, err := setup(t, tc.preset)
			if err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			for i, row := range tc.rows {
				p, err := result.GetProfile(row.profile)
				if err != nil {
					t.Errorf("row_%d: result.GetProfile: %q", i, err)
					continue
				}
				g, err := result.GetGroup(row.group)
				if err != nil {
					t.Errorf("row_%d: result.GetGroup: %q", i, err)
					continue
				}
				err = s.DeleteGroup(g.Id, p.Id)
				if !errors.Is(err, row.wantErr) {
					t.Errorf("row_%d: expected error: %q, got: %q", i, row.wantErr, err)
				}
			}
		})
	}
}

type PresetResult struct {
	profiles map[string]*model.Profile
	groups   map[string]*model.Group
}

func (result *PresetResult) GetGroup(key string) (*model.Group, error) {
	if group, ok := result.groups[key]; ok {
		return group, nil
	}
	return nil, fmt.Errorf("group key %q not found", key)
}

func (result *PresetResult) GetProfile(key string) (*model.Profile, error) {
	if profile, ok := result.profiles[key]; ok {
		return profile, nil
	}
	return nil, fmt.Errorf("profile key %q not found", key)
}

type Preset struct {
	profiles map[string]presetProfile
	groups   map[string]presetGroup
}

type presetProfile struct {
	userId int
	name   string
}

type presetGroup struct {
	name    string
	owner   string
	manager []string
	member  []string
}

func NewPresetResult() *PresetResult {
	var result PresetResult
	result.profiles = make(map[string]*model.Profile)
	result.groups = make(map[string]*model.Group)
	return &result
}

func setup(t *testing.T, preset *Preset) (*ContactsService, *PresetResult, error) {
	t.Helper()
	result := NewPresetResult()
	s, err := NewTestContactsService()
	if err != nil {
		t.Fatalf("NewTestContactsService failed: %v", err)
	}

	if preset != nil {
		for key, prep := range preset.profiles {
			p, err := s.CreateProfile(prep.userId, prep.name)
			if err != nil {
				return nil, nil, fmt.Errorf("presetProfile.CreateProfile: %q", err)
			}
			result.profiles[key] = p
		}

		for key, preg := range preset.groups {
			owner, err := result.GetProfile(preg.owner)
			if err != nil {
				return nil, nil, fmt.Errorf("presetGroup.result.GetProfile: %q", err)
			}
			g, err := s.CreateGroup(owner.Id, preg.name)
			if err != nil {
				return nil, nil, fmt.Errorf("presetGroup.CreateGroup: %q", err)
			}
			result.groups[key] = g

			for _, key := range preg.member {
				mbrProfile, err := result.GetProfile(key)
				if err != nil {
					return nil, nil, fmt.Errorf("presetGroupMember.result.GetProfile: %q", err)

				}
				if _, err = s.Invite(g.Id, owner.Id, mbrProfile.Id); err != nil {
					return nil, nil, fmt.Errorf("presetGroup.member.Invite: %q", err)
				}
			}
		}

	}

	return s, result, nil
}
