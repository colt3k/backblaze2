package backblaze2

import (
	"github.com/colt3k/backblaze2/b2api"
	"github.com/colt3k/backblaze2/internal/auth"
)

func Authorize(authConfig b2api.AuthConfig) (*b2api.AuthorizationResp, error) {
	authd, err := auth.AuthorizeAccount(authConfig)

	if err != nil {
		return nil, err
	}
	return authd, nil
}