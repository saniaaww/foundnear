
package middleware

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTClaims(t *testing.T) {
	claims := JWTClaims{
		UserID: 1,
		Role:   "user",
	}

	if claims.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", claims.UserID)
	}

	if claims.Role != "user" {
		t.Errorf("expected Role user, got %s", claims.Role)
	}
}

func TestJWTSigningMethod(t *testing.T) {
	token := jwt.New(jwt.SigningMethodHS256)

	if token.Method.Alg() != "HS256" {
		t.Errorf("expected HS256, got %s", token.Method.Alg())
	}
}
