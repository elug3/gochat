package httpclient

import "context"

type Api struct {
	Contacts *ContactsClient
	Messages *MessagesClient
}

func NewApi(opts ...RequestOption) *Api {
	return &Api{
		Contacts: NewContactsClient(opts...),
		Messages: NewMessagesClient(opts...),
	}
}

func (api *Api) IsAuthenticated(ctx context.Context) (bool, error) {
	return true, nil
}
