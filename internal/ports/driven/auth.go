package driven

import (
	"context"
	"time"
)

type AuthProviderPort interface {
	Name() string
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (OAuthToken, error)
	GetUserInfo(ctx context.Context, token OAuthToken) (OAuthUser, error)
}

type OAuthToken struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	Scopes       []string
	TokenType    string
}

type OAuthUser struct {
	ID         string
	Provider   string
	Email      string
	Name       string
	AvatarURL  string
	RawPayload map[string]interface{}
}

type SocialitePort interface {
	Provider(name string) (AuthProviderPort, error)
	Register(name string, provider AuthProviderPort)
	Drivers() []string
}
