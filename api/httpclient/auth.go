package httpclient

type AuthClient struct {
	options []RequestOption
}

func NewAuthClient(opts ...RequestOption) *AuthClient {
	return &AuthClient{
		options: opts,
	}
}
