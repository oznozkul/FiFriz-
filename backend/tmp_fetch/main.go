package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type claims struct {
	UserID   string `json:"userId"`
	TenantID string `json:"tenantId"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func main() {
	now := time.Now()
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		UserID: "064c6909-f79f-4625-8961-fa1c2f33aebc", TenantID: "e41d662f-2dfb-474f-9b18-83369c17aa25",
		Email: "aliozanozkul96@gmail.com", Role: "owner",
		RegisteredClaims: jwt.RegisteredClaims{Issuer: "fifriz", Subject: "064c6909-f79f-4625-8961-fa1c2f33aebc",
			IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour))},
	}).SignedString([]byte("dev-secret-change-me"))

	req, _ := http.NewRequest(http.MethodGet, "http://localhost:3000/api/v1/documents/feee37db-4855-42fb-b2b9-04e089da9e1d/file", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("status=%d bytes=%d\n", resp.StatusCode, len(body))
}
