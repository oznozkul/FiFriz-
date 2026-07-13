package integrations

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const credentialEnvelopeVersion = 1

type oauthCredentialEnvelope struct {
	Version      int    `json:"version"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
	Expiry       string `json:"expiry"`
}

type OAuthTokens struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	Expiry       time.Time
}

type credentialEnvelope struct {
	Version  int    `json:"version"`
	Password string `json:"password"`
}

// CredentialService centralizes integration secret access.
// Future versions can replace base64 storage with real encryption.
type CredentialService struct{}

func NewCredentialService() *CredentialService {
	return &CredentialService{}
}

func (s *CredentialService) StorePassword(password string) ([]byte, error) {
	password = strings.TrimSpace(password)
	if password == "" {
		return nil, errors.New("password is required")
	}

	envelope := credentialEnvelope{
		Version:  credentialEnvelopeVersion,
		Password: password,
	}
	return json.Marshal(envelope)
}

func (s *CredentialService) GetPassword(encrypted []byte) (string, error) {
	if len(encrypted) == 0 {
		return "", errors.New("credentials are not configured")
	}

	var envelope credentialEnvelope
	if err := json.Unmarshal(encrypted, &envelope); err == nil && envelope.Password != "" {
		return envelope.Password, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(string(encrypted))
	if err != nil {
		return "", errors.New("credentials could not be decoded")
	}
	if err := json.Unmarshal(decoded, &envelope); err != nil {
		return "", errors.New("credentials could not be parsed")
	}
	if strings.TrimSpace(envelope.Password) == "" {
		return "", errors.New("password is missing in credentials")
	}
	return envelope.Password, nil
}

func (s *CredentialService) StoreOAuthTokens(tokens OAuthTokens) ([]byte, error) {
	if strings.TrimSpace(tokens.AccessToken) == "" {
		return nil, errors.New("access token is required")
	}
	expiry := ""
	if !tokens.Expiry.IsZero() {
		expiry = tokens.Expiry.UTC().Format(time.RFC3339)
	}
	envelope := oauthCredentialEnvelope{
		Version:      credentialEnvelopeVersion,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		Expiry:       expiry,
	}
	return json.Marshal(envelope)
}

func (s *CredentialService) GetOAuthTokens(encrypted []byte) (OAuthTokens, error) {
	if len(encrypted) == 0 {
		return OAuthTokens{}, errors.New("credentials are not configured")
	}

	var oauthEnvelope oauthCredentialEnvelope
	if err := json.Unmarshal(encrypted, &oauthEnvelope); err == nil && oauthEnvelope.AccessToken != "" {
		return parseOAuthEnvelope(oauthEnvelope)
	}

	var envelope credentialEnvelope
	if err := json.Unmarshal(encrypted, &envelope); err == nil && envelope.Password != "" {
		return OAuthTokens{}, errors.New("oauth tokens are not configured")
	}

	decoded, err := base64.StdEncoding.DecodeString(string(encrypted))
	if err != nil {
		return OAuthTokens{}, errors.New("credentials could not be decoded")
	}
	if err := json.Unmarshal(decoded, &oauthEnvelope); err == nil && oauthEnvelope.AccessToken != "" {
		return parseOAuthEnvelope(oauthEnvelope)
	}
	return OAuthTokens{}, errors.New("oauth tokens could not be parsed")
}

func parseOAuthEnvelope(envelope oauthCredentialEnvelope) (OAuthTokens, error) {
	tokens := OAuthTokens{
		AccessToken:  strings.TrimSpace(envelope.AccessToken),
		RefreshToken: strings.TrimSpace(envelope.RefreshToken),
		TokenType:    strings.TrimSpace(envelope.TokenType),
	}
	if tokens.AccessToken == "" {
		return OAuthTokens{}, errors.New("access token is missing")
	}
	if envelope.Expiry != "" {
		parsed, err := time.Parse(time.RFC3339, envelope.Expiry)
		if err == nil {
			tokens.Expiry = parsed.UTC()
		}
	}
	return tokens, nil
}

func SanitizeSettings(settings map[string]any) map[string]any {
	if len(settings) == 0 {
		return settings
	}
	clean := make(map[string]any, len(settings))
	for key, value := range settings {
		lower := strings.ToLower(strings.TrimSpace(key))
		if lower == "password" || lower == "secret" || lower == "token" {
			continue
		}
		clean[key] = value
	}
	return clean
}
