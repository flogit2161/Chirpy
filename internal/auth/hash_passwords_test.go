package auth

import (
	"testing"
)

func TestHashing(t *testing.T) {
	password := "password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPW function errored, error :%v", err)
	}

	if hash == "" {
		t.Errorf("HashPW returned an empty string")
	}

	if hash == "password" {
		t.Errorf("HashPW returned the same pw")
	}
}

func TestCheckHash(t *testing.T) {
	password := "password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPW function errored, error:%v", err)
	}

	match, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("CheckPwHash function errored, error:%v", err)
	}

	if match == false {
		t.Errorf("Match returned false on password check")
	}
}

func TestWrongPWHash(t *testing.T) {
	password := "password"
	wrongPassword := "passwordtest"
	hash, err := HashPassword(password)

	wrongMatch, err := CheckPasswordHash(wrongPassword, hash)
	if err != nil {
		t.Fatalf("CheckPwHash function errored, error:%v", err)
	}

	if wrongMatch == true {
		t.Errorf("Match returned true, expected to be false with wrong password")
	}
}
