package webauthn

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/rs/zerolog/log"
)

var (
	session = struct {
		mu    sync.Mutex
		store map[string]*webauthn.SessionData
	}{
		store: make(map[string]*webauthn.SessionData),
	}
	users = map[string]*User{}
)

type Handler struct {
	wAuth *webauthn.WebAuthn
}

func NewHandler(opts *Options) (*Handler, error) {
	wAuth, err := webauthn.New(&webauthn.Config{
		RPDisplayName: opts.RPDisplayName,
		RPID:          opts.RPID,
		RPOrigins:     opts.RPOrigins,
	})
	if err != nil {
		return nil, err
	}
	return &Handler{wAuth: wAuth}, nil
}

func (h *Handler) registerStart(w http.ResponseWriter, r *http.Request) {
	userId := "1"

	u := &User{
		Id:          []byte(userId),
		Name:        "testing",
		DisplayName: "Testing User",
		Creds:       []webauthn.Credential{},
	}
	var (
		creation *protocol.CredentialCreation
		s        *webauthn.SessionData
		err      error
	)
	opts := []webauthn.RegistrationOption{
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		webauthn.WithExclusions(webauthn.Credentials(u.WebAuthnCredentials()).CredentialDescriptors()),
		webauthn.WithExtensions(map[string]any{"credProps": true}),
	}
	if creation, s, err = h.wAuth.BeginMediatedRegistration(u, protocol.MediationDefault, opts...); err != nil {
		http.Error(w, "Failed to begin registration", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Registration error")
		return
	}

	saveSession(userId, s)
	saveUser(userId, u)

	if err = json.NewEncoder(w).Encode(creation); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Encoding error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) registerFinish(w http.ResponseWriter, r *http.Request) {
	userId := "1"
	u, ok := loadUser(userId)
	if !ok {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}
	sess, ok := loadSession(userId)
	if !ok {
		http.Error(w, "Session not found", http.StatusBadRequest)
		return
	}

	cred, err := h.wAuth.FinishRegistration(u, *sess, r)
	if err != nil {
		http.Error(w, "Failed to finish registration", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Registration error")
		return
	}
	u.Creds = append(u.Creds, *cred)

	w.Write([]byte("Registration successful"))
}

func (h *Handler) loginStart(w http.ResponseWriter, r *http.Request) {
	assertion, s, err := h.wAuth.BeginDiscoverableMediatedLogin(protocol.MediationDefault)
	if err != nil {
		http.Error(w, "Failed to begin login", http.StatusInternalServerError)
		return
	}
	userId := "1"
	saveSession(userId, s)

	if err = json.NewEncoder(w).Encode(assertion); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Encoding error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) loginFinish(w http.ResponseWriter, r *http.Request) {
	userId := "1"
	s, ok := loadSession(userId)
	if !ok {
		http.Error(w, "Session not found", http.StatusBadRequest)
		return
	}
	validatedUser, validatedCredential, err := h.wAuth.FinishPasskeyLogin(func(rawID, userHandle []byte) (user webauthn.User, err error) {
		u, ok := loadUser(userId)
		if !ok {
			return nil, fmt.Errorf("User not found")
		}
		return u, nil
	}, *s, r)
	if err != nil {
		http.Error(w, "Failed to finish login", http.StatusInternalServerError)
		log.Error().Err(err).Msg("Login error")
		return
	}
	user, ok := validatedUser.(*User)
	if !ok {
		http.Error(w, "Failed to assert user", http.StatusInternalServerError)
		log.Error().Msg("Failed to assert user")
		return
	}

	var found bool
	for i, credential := range user.Creds {
		if bytes.Equal(validatedCredential.ID, credential.ID) {
			user.Creds[i] = *validatedCredential

			saveUser(userId, user)
			found = true
			break
		}
	}
	if !found {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func saveSession(key string, sess *webauthn.SessionData) {
	session.mu.Lock()
	defer session.mu.Unlock()
	session.store[key] = sess
}

func loadSession(key string) (*webauthn.SessionData, bool) {
	session.mu.Lock()
	defer session.mu.Unlock()
	sess, ok := session.store[key]
	return sess, ok
}

func deleteSession(key string) {
	session.mu.Lock()
	defer session.mu.Unlock()
	delete(session.store, key)
}

func saveUser(key string, u *User) {
	users[key] = u
}

func loadUser(key string) (*User, bool) {
	u, ok := users[key]
	return u, ok
}

func (h *Handler) registeredCredentials(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(session.store)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
