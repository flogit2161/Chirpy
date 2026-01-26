package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWT(t *testing.T) {
	userID := uuid.New()
	jwt, err := MakeJWT(userID, "secretID", time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT function errored")
	}

	validate, err := ValidateJWT(jwt, "secretID")
	if err != nil {
		t.Fatalf("ValidateJWT function errored")
	}

	if userID != validate {
		t.Errorf("Error validating token, UserID : %v ValidatedID : %v", userID, validate)
	}

}

func TestJWTWrongSecret(t *testing.T) {
	userID := uuid.New()
	jwt, err := MakeJWT(userID, "secretID", time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT function errored")
	}

	validate, err := ValidateJWT(jwt, "wrongsecret")
	if err == nil {
		t.Fatalf("Validate JWT validated when using wrong secret. UserID : %v ValidatedID : %v", userID, validate)
	}
}

func TestJWTExpiredToken(t *testing.T) {
	userID := uuid.New()
	token, err := MakeJWT(userID, "secretID", -time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT errored")
	}

	validate, err := ValidateJWT(token, "secret")
	if err == nil {
		t.Fatalf("Validate JWT validated when token expired. UserID :%v ValidatedID :%v", userID, validate)
	}
}
