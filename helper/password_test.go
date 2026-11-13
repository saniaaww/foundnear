
package helper

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	password := "password123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if hash == password {
		t.Fatal("password should not be stored as plain text")
	}

	if !CheckPassword(password, hash) {
		t.Fatal("password verification should succeed")
	}

	if CheckPassword("wrongpassword", hash) {
		t.Fatal("wrong password should not be accepted")
	}
}
