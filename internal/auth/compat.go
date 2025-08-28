package auth

import (
	"fmt"

	"github.com/sstent/go-garminconnect/internal/auth/garth"
)

// LegacyAuthToGarth converts a legacy authentication token to a garth session
func LegacyAuthToGarth(legacyToken *Token) (*garth.Session, error) {
	if legacyToken == nil {
		return nil, fmt.Errorf("legacy token cannot be nil")
	}

	return &garth.Session{
		OAuth1Token:  legacyToken.OAuthToken,
		OAuth1Secret: legacyToken.OAuthSecret,
		OAuth2Token:  legacyToken.AccessToken,
	}, nil
}

// GarthToLegacyAuth converts a garth session to a legacy authentication token
func GarthToLegacyAuth(session *garth.Session) (*Token, error) {
	if session == nil {
		return nil, fmt.Errorf("session cannot be nil")
	}

	return &Token{
		OAuthToken:  session.OAuth1Token,
		OAuthSecret: session.OAuth1Secret,
		AccessToken: session.OAuth2Token,
	}, nil
}
