package jwk

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"os"

	"github.com/lestrrat-go/jwx/v3/jwk"
)

type Jwks struct {
	set jwk.Set
}

func NewJwks() *Jwks {
	set := jwk.NewSet()
	return &Jwks{
		set: set,
	}
}

func (j *Jwks) AddKey(key rsa.PublicKey, kid string) error {
	jwkKey, err := jwk.Import(key)
	if err != nil {
		return err
	}
	jwkKey.Set(jwk.KeyIDKey, kid)
	jwkKey.Set(jwk.AlgorithmKey, "RS256")
	jwkKey.Set(jwk.KeyUsageKey, "sig")

	j.set.AddKey(jwkKey)
	return nil
}

func LoadFile(path string) (*Jwks, error) {
	jwksData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	set, err := jwk.Parse(jwksData)
	if err != nil {
		return nil, err
	}

	jwks := Jwks{
		set: set,
	}
	return &jwks, nil
}

func (j *Jwks) Json() ([]byte, error) {
	if j.set == nil {
		return nil, fmt.Errorf("jwks set not initialized")
	}
	return json.MarshalIndent(j.set, "", "  ")
}
