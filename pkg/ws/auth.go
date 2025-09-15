package ws

import (
	"context"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/patrickmn/go-cache"
)

type JwkAuthenticator struct {
	jwksURL string
	cache   *cache.Cache
}

func NewAuthenticator(authUrl string) (*JwkAuthenticator, error) {
	jwksURL := authUrl + "/.well-known/jwks.json"
	return &JwkAuthenticator{
		jwksURL: jwksURL,
		cache:   cache.New(5*time.Minute, 10*time.Minute),
	}, nil
}

func (auth *JwkAuthenticator) Authenticate(ctx context.Context, token string) (int32, error) {
	set, err := auth.getJWKS(ctx)
	if err != nil {
		return 0, err
	}
	_ = set
	return 0, nil
}

func (auth *JwkAuthenticator) getJWKS(ctx context.Context) (jwk.Set, error) {
	return nil, nil
}

func fetchJWKS(ctx context.Context, url string) (jwk.Set, error) {
	set, err := jwk.Fetch(ctx, url)
	if err != nil {
		return nil, err
	}

	return set, nil
}
