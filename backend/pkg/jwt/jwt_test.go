package jwt

import (
	"testing"
	"time"
)

func TestGenerateAndValidateToken(t *testing.T) {
	secret := "test-secret-key"
	username := "admin"
	expiryHours := 24

	// Generate token
	token, err := GenerateToken(username, secret, expiryHours)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if token == "" {
		t.Error("GenerateToken() returned empty token")
	}

	// Validate token
	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if claims.Username != username {
		t.Errorf("Username = %v, want %v", claims.Username, username)
	}
}

func TestValidateToken_InvalidSecret(t *testing.T) {
	secret := "correct-secret"
	wrongSecret := "wrong-secret"

	token, _ := GenerateToken("admin", secret, 24)

	_, err := ValidateToken(token, wrongSecret)
	if err == nil {
		t.Error("ValidateToken() should fail with wrong secret")
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	secret := "test-secret"

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "invalid format",
			token: "invalid.token.format",
		},
		{
			name:  "random string",
			token: "not-a-valid-jwt-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateToken(tt.token, secret)
			if err == nil {
				t.Error("ValidateToken() should fail with invalid token")
			}
		})
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	secret := "test-secret"

	// Generate token with negative expiry (already expired)
	token, _ := GenerateToken("admin", secret, -1)

	// Wait a moment to ensure token is expired
	time.Sleep(time.Millisecond * 100)

	_, err := ValidateToken(token, secret)
	if err == nil {
		t.Error("ValidateToken() should fail with expired token")
	}
}

func TestTokenIssuer(t *testing.T) {
	secret := "test-secret"
	token, _ := GenerateToken("admin", secret, 24)

	claims, _ := ValidateToken(token, secret)
	if claims.Issuer != "smpp-simulator" {
		t.Errorf("Issuer = %v, want smpp-simulator", claims.Issuer)
	}
}
