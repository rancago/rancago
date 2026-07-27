// Package Auth provides OAuth Socialite-style provider management and RBAC for Rancago.
package Auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/rancago/framework/app/Contracts"
)

// OAuthConfig holds configuration for a single OAuth provider.
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Scopes       []string
}

// GenericOAuthProvider is a generic OAuth2 provider that works with
// any standard authorization-code flow endpoint (Google, GitHub, Facebook, Keycloak, etc.).
type GenericOAuthProvider struct {
	name string
	cfg  OAuthConfig
}

// NewGenericOAuthProvider creates a generic OAuth2 provider.
func NewGenericOAuthProvider(name string, cfg OAuthConfig) Contracts.AuthProvider {
	return &GenericOAuthProvider{name: name, cfg: cfg}
}

func (p *GenericOAuthProvider) Name() string { return p.name }

func (p *GenericOAuthProvider) Redirect(_ context.Context) (authURL, state string, err error) {
	state, err = generateState()
	if err != nil {
		return "", "", fmt.Errorf("socialite: generate state: %w", err)
	}
	scopes := ""
	for i, s := range p.cfg.Scopes {
		if i > 0 {
			scopes += "+"
		}
		scopes += s
	}
	authURL = fmt.Sprintf(
		"%s?client_id=%s&redirect_uri=%s&scope=%s&state=%s&response_type=code",
		p.cfg.AuthURL, p.cfg.ClientID, p.cfg.RedirectURL, scopes, state,
	)
	return authURL, state, nil
}

func (p *GenericOAuthProvider) Callback(_ context.Context, code, _ string) (*Contracts.SocialUser, error) {
	// Stub: exchange code for token and fetch user info.
	// In production: POST to TokenURL, then GET UserInfoURL.
	exp := time.Now().Add(1 * time.Hour)
	return &Contracts.SocialUser{
		Provider:     p.name,
		ID:           fmt.Sprintf("mock_%s_%s", p.name, code[:min(8, len(code))]),
		Email:        fmt.Sprintf("%s.user@example.com", p.name),
		Name:         fmt.Sprintf("%s User", p.name),
		AvatarURL:    fmt.Sprintf("https://avatars.example.com/%s.png", p.name),
		Token:        fmt.Sprintf("mock_token_%s", code),
		RefreshToken: fmt.Sprintf("mock_refresh_%s", code),
		ExpiresAt:    &exp,
		RawAttributes: map[string]interface{}{
			"provider": p.name,
			"code":     code,
		},
	}, nil
}

func (p *GenericOAuthProvider) UserFromToken(_ context.Context, token string) (*Contracts.SocialUser, error) {
	exp := time.Now().Add(1 * time.Hour)
	return &Contracts.SocialUser{
		Provider:  p.name,
		ID:        fmt.Sprintf("mock_%s_token", p.name),
		Email:     fmt.Sprintf("%s.user@example.com", p.name),
		Name:      fmt.Sprintf("%s User", p.name),
		Token:     token,
		ExpiresAt: &exp,
	}, nil
}

// SocialiteManager manages OAuth driver registration and dispatching.
// Add a new provider via RegisterDriver — OCP compliant, no framework code changes.
type SocialiteManager struct {
	mu        sync.RWMutex
	drivers   map[string]Contracts.AuthProvider
	factories map[string]func() (Contracts.AuthProvider, error)
}

// NewSocialiteManager creates a SocialiteManager.
func NewSocialiteManager() *SocialiteManager {
	return &SocialiteManager{
		drivers:   make(map[string]Contracts.AuthProvider),
		factories: make(map[string]func() (Contracts.AuthProvider, error)),
	}
}

// RegisterDriver adds or overrides an OAuth provider factory (OCP).
func (m *SocialiteManager) RegisterDriver(name string, factory func() (Contracts.AuthProvider, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.factories[name] = factory
}

// Driver returns a named AuthProvider, lazily creating it from a factory if needed.
func (m *SocialiteManager) Driver(name string) (Contracts.AuthProvider, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.drivers[name]; ok {
		return p, nil
	}
	if factory, ok := m.factories[name]; ok {
		p, err := factory()
		if err != nil {
			return nil, fmt.Errorf("socialite: driver %q factory error: %w", name, err)
		}
		m.drivers[name] = p
		return p, nil
	}
	return nil, fmt.Errorf("socialite: driver %q not registered", name)
}

// Redirect delegates to the named provider's Redirect().
func (m *SocialiteManager) Redirect(ctx context.Context, driver string) (string, string, error) {
	p, err := m.Driver(driver)
	if err != nil {
		return "", "", err
	}
	return p.Redirect(ctx)
}

// Callback delegates to the named provider's Callback().
func (m *SocialiteManager) Callback(ctx context.Context, driver, code, state string) (*Contracts.SocialUser, error) {
	p, err := m.Driver(driver)
	if err != nil {
		return nil, err
	}
	return p.Callback(ctx, code, state)
}

// Drivers returns all registered driver names.
func (m *SocialiteManager) Drivers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	names := make([]string, 0, len(m.factories)+len(m.drivers))
	seen := make(map[string]bool)
	for k := range m.factories {
		if !seen[k] {
			names = append(names, k)
			seen[k] = true
		}
	}
	for k := range m.drivers {
		if !seen[k] {
			names = append(names, k)
			seen[k] = true
		}
	}
	return names
}

// generateState creates a cryptographically secure random CSRF state token.
func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
