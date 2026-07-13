package sharing

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const shareAccessTokenDuration = 30 * time.Minute

type ShareAccessClaims struct {
	ShareLinkID string `json:"shareLinkId"`
	ShareToken  string `json:"shareToken"`
	jwt.RegisteredClaims
}

func issueShareAccessToken(secret, shareLinkID, shareToken string) (string, int, error) {
	now := time.Now()
	expiresAt := now.Add(shareAccessTokenDuration)
	claims := ShareAccessClaims{
		ShareLinkID: shareLinkID,
		ShareToken:  shareToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "fifriz-share",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", 0, err
	}

	return signed, int(shareAccessTokenDuration.Seconds()), nil
}

func parseShareAccessToken(secret, accessToken, expectedShareToken string) error {
	if accessToken == "" {
		return errors.New("access token required")
	}

	parsed, err := jwt.ParseWithClaims(accessToken, &ShareAccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return err
	}

	claims, ok := parsed.Claims.(*ShareAccessClaims)
	if !ok || !parsed.Valid {
		return errors.New("invalid access token")
	}
	if claims.ShareToken != expectedShareToken {
		return errors.New("access token mismatch")
	}

	return nil
}
