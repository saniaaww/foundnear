
package helper

import "testing"

func TestValidateRegister(t *testing.T) {
	err := ValidateRegister("Sania", "sania@example.com", "password123")
	if err != nil {
		t.Fatalf("expected valid register data, got: %v", err)
	}

	err = ValidateRegister("", "sania@example.com", "password123")
	if err == nil {
		t.Fatal("expected name validation error")
	}
}

func TestValidateLogin(t *testing.T) {
	err := ValidateLogin("sania@example.com", "password123")
	if err != nil {
		t.Fatalf("expected valid login data, got: %v", err)
	}

	err = ValidateLogin("invalid-email", "password123")
	if err == nil {
		t.Fatal("expected email validation error")
	}
}
