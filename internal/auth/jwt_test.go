package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTCreationAndValidation(t *testing.T) {
	userID := uuid.New()
	secret := "supersecret"
	expiresIn := time.Hour

	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	returnedID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if returnedID != userID {
		t.Errorf("expected userID %v, got %v", userID, returnedID)
	}
}
