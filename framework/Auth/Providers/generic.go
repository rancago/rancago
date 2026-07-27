// Package Providers contains OAuth provider implementations for Rancago.
// Each provider satisfies Contracts.AuthProvider.
// Register new providers via SocialiteManager.RegisterDriver (OCP).
package Providers

import (
	"github.com/rancago/framework/app/Contracts"
	"github.com/rancago/framework/framework/Auth"
)

// NewGoogleProvider creates a pre-configured Google OAuth provider.
func NewGoogleProvider(clientID, clientSecret, redirectURL string) Contracts.AuthProvider {
	return Auth.NewGenericOAuthProvider("google", Auth.OAuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		AuthURL:      "https://accounts.google.com/o/oauth2/auth",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://www.googleapis.com/oauth2/v3/userinfo",
		Scopes:       []string{"email", "profile"},
	})
}

// NewGitHubProvider creates a pre-configured GitHub OAuth provider.
func NewGitHubProvider(clientID, clientSecret, redirectURL string) Contracts.AuthProvider {
	return Auth.NewGenericOAuthProvider("github", Auth.OAuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		AuthURL:      "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		UserInfoURL:  "https://api.github.com/user",
		Scopes:       []string{"user:email", "read:user"},
	})
}

// NewFacebookProvider creates a pre-configured Facebook OAuth provider.
func NewFacebookProvider(clientID, clientSecret, redirectURL string) Contracts.AuthProvider {
	return Auth.NewGenericOAuthProvider("facebook", Auth.OAuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		AuthURL:      "https://www.facebook.com/v18.0/dialog/oauth",
		TokenURL:     "https://graph.facebook.com/v18.0/oauth/access_token",
		UserInfoURL:  "https://graph.facebook.com/me?fields=id,name,email",
		Scopes:       []string{"email", "public_profile"},
	})
}
