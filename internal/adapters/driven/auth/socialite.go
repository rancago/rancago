package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/rancago/framework/internal/kernel"
	"github.com/rancago/framework/internal/ports/driven"
)

type GenericOAuthProvider struct {
	name   string
	cfg    kernel.OAuthProviderConfig
}

func NewGenericOAuthProvider(name string, cfg kernel.OAuthProviderConfig) driven.AuthProviderPort {
	return &GenericOAuthProvider{name: name, cfg: cfg}
}

func (p *GenericOAuthProvider) Name() string { return p.name }

func (p *GenericOAuthProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&scope=%s&state=%s&response_type=code",
		p.cfg.AuthURL, p.cfg.ClientID, p.cfg.RedirectURL, joinScopes(p.cfg.Scopes), state)
}

func joinScopes(scopes []string) string {
	out := ""
	for i, s := range scopes {
		if i > 0 {
			out += " "
		}
		out += s
	}
	return out
}

func (p *GenericOAuthProvider) ExchangeCode(_ context.Context, code string) (driven.OAuthToken, error) {
	return driven.OAuthToken{
		AccessToken:  fmt.Sprintf("mock_access_%s_%d", code, time.Now().Unix()),
		RefreshToken: fmt.Sprintf("mock_refresh_%s", code),
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		TokenType:    "Bearer",
		Scopes:       p.cfg.Scopes,
	}, nil
}

func (p *GenericOAuthProvider) GetUserInfo(_ context.Context, token driven.OAuthToken) (driven.OAuthUser, error) {
	return driven.OAuthUser{
		ID:         fmt.Sprintf("mock_user_%s", p.name),
		Provider:   p.name,
		Email:      fmt.Sprintf("%s.user@example.com", p.name),
		Name:       fmt.Sprintf("%s User", p.name),
		AvatarURL:  fmt.Sprintf("https://avatars.example.com/%s/mock.png", p.name),
		RawPayload: map[string]interface{}{"access_token": token.AccessToken},
	}, nil
}

type SocialiteManager struct {
	cfg       *kernel.AuthConfig
	providers map[string]driven.AuthProviderPort
}

func NewSocialiteManager(cfg *kernel.AuthConfig) driven.SocialitePort {
	m := &SocialiteManager{
		cfg:       cfg,
		providers: make(map[string]driven.AuthProviderPort),
	}
	for name, p := range cfg.Providers {
		m.providers[name] = NewGenericOAuthProvider(name, p)
	}
	return m
}

func (s *SocialiteManager) Provider(name string) (driven.AuthProviderPort, error) {
	p, ok := s.providers[name]
	if !ok {
		return nil, fmt.Errorf("socialite: provider %s not registered", name)
	}
	return p, nil
}

func (s *SocialiteManager) Register(name string, provider driven.AuthProviderPort) {
	s.providers[name] = provider
}

func (s *SocialiteManager) Drivers() []string {
	out := make([]string, 0, len(s.providers))
	for k := range s.providers {
		out = append(out, k)
	}
	return out
}
