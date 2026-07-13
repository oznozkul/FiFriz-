package integrations

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const oauthStateTTL = 10 * time.Minute

type OAuthStatePayload struct {
	TenantID  uuid.UUID `json:"tenantId"`
	UserID    uuid.UUID `json:"userId"`
	Nonce     string    `json:"nonce"`
	ExpiresAt int64     `json:"exp"`
}

func newOAuthState(tenantID, userID uuid.UUID) (OAuthStatePayload, error) {
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return OAuthStatePayload{}, err
	}
	return OAuthStatePayload{
		TenantID:  tenantID,
		UserID:    userID,
		Nonce:     base64.RawURLEncoding.EncodeToString(nonceBytes),
		ExpiresAt: time.Now().UTC().Add(oauthStateTTL).Unix(),
	}, nil
}

func signOAuthState(secret string, payload OAuthStatePayload) (string, error) {
	if strings.TrimSpace(secret) == "" {
		return "", errors.New("oauth state secret is not configured")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(raw)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(encoded))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%s.%s", encoded, signature), nil
}

func verifyOAuthState(secret, state string) (OAuthStatePayload, error) {
	parts := strings.Split(state, ".")
	if len(parts) != 2 {
		return OAuthStatePayload{}, errors.New("invalid oauth state")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(parts[0]))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return OAuthStatePayload{}, errors.New("invalid oauth state signature")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return OAuthStatePayload{}, errors.New("invalid oauth state payload")
	}
	var payload OAuthStatePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return OAuthStatePayload{}, errors.New("invalid oauth state payload")
	}
	if time.Now().UTC().Unix() > payload.ExpiresAt {
		return OAuthStatePayload{}, errors.New("oauth state expired")
	}
	return payload, nil
}
