package Providers

import (
	"github.com/rancago/framework/app/Contracts"
	"github.com/rancago/framework/framework/Auth"
	authProviders "github.com/rancago/framework/framework/Auth/Providers"
	"github.com/rancago/framework/framework/Cache"
	"github.com/rancago/framework/framework/Container"
)

// AuthServiceProvider registers and boots the Socialite OAuth manager and RBAC service.
//
// Container bindings (post-registration):
//   - "auth.socialite"            → *Auth.SocialiteManager (singleton)
//   - "Contracts.SocialiteManager" → alias
//   - "auth.rbac"                 → Contracts.RBACService (singleton)
//   - "Contracts.RBACService"     → alias
type AuthServiceProvider struct {
	oauthCfgs   map[string]Auth.OAuthConfig
	rbacUserKey interface{}
}

// NewAuthServiceProvider creates an AuthServiceProvider.
// oauthCfgs maps provider names to their OAuth config (google, github, facebook, etc.).
// rbacUserKey is the context key used to extract the authenticated user ID in RBAC middleware.
func NewAuthServiceProvider(oauthCfgs map[string]Auth.OAuthConfig, rbacUserKey interface{}) *AuthServiceProvider {
	return &AuthServiceProvider{oauthCfgs: oauthCfgs, rbacUserKey: rbacUserKey}
}

// Register binds the Socialite and RBAC services into the container.
func (p *AuthServiceProvider) Register(c *Container.Container) error {
	c.Singleton("auth.socialite", func(c *Container.Container) (interface{}, error) {
		mgr := Auth.NewSocialiteManager()
		// Register built-in OAuth providers.
		for name, cfg := range p.oauthCfgs {
			providerCfg := cfg // capture loop variable
			mgr.RegisterDriver(name, func() (Contracts.AuthProvider, error) {
				return Auth.NewGenericOAuthProvider(name, providerCfg), nil
			})
		}
		// Named factories for the 3 pre-configured providers.
		if _, ok := p.oauthCfgs["google"]; ok {
			gcfg := p.oauthCfgs["google"]
			mgr.RegisterDriver("google", func() (Contracts.AuthProvider, error) {
				return authProviders.NewGoogleProvider(gcfg.ClientID, gcfg.ClientSecret, gcfg.RedirectURL), nil
			})
		}
		if _, ok := p.oauthCfgs["github"]; ok {
			ghcfg := p.oauthCfgs["github"]
			mgr.RegisterDriver("github", func() (Contracts.AuthProvider, error) {
				return authProviders.NewGitHubProvider(ghcfg.ClientID, ghcfg.ClientSecret, ghcfg.RedirectURL), nil
			})
		}
		if _, ok := p.oauthCfgs["facebook"]; ok {
			fbcfg := p.oauthCfgs["facebook"]
			mgr.RegisterDriver("facebook", func() (Contracts.AuthProvider, error) {
				return authProviders.NewFacebookProvider(fbcfg.ClientID, fbcfg.ClientSecret, fbcfg.RedirectURL), nil
			})
		}
		return mgr, nil
	})
	c.Alias("auth.socialite", "Contracts.SocialiteManager")

	c.Singleton("auth.rbac", func(c *Container.Container) (interface{}, error) {
		redisRaw, err := c.Resolve("redis")
		if err != nil {
			return nil, err
		}
		redis := redisRaw.(*Cache.RedisManager)
		return Auth.NewRBACService(redis, p.rbacUserKey), nil
	})
	c.Alias("auth.rbac", "Contracts.RBACService")
	return nil
}

// Boot is a no-op for auth - all wiring happens in Register.
func (p *AuthServiceProvider) Boot(_ *Container.Container) error { return nil }
