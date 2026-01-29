package auth

import (
	"net/http"
	"testing"
)

func TestBearerToken(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer 123")

	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("BearerToken function errored")
	}

	if token != "123" {
		t.Errorf("Error cleaning header. Expected : 123, Token : %v", token)
	}
}

func TestBearerTokenEmpty(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "")

	token, err := GetBearerToken(headers)
	if err == nil {
		t.Fatalf("BearerToken function did not error on empty token. Token : %v", token)
	}

}

func TestBearerTokenTrim(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer     123")

	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("BearerToken function errored")
	}

	if token != "123" {
		t.Errorf("Error cleaning header. Expected : 123, Token : %v", token)
	}
}

func TestMakeRefreshToken(t *testing.T) {
	encodedHex, err := MakeRefreshToken()
	if err != nil {
		t.Fatalf("MakeRefreshToken function errored")
	}

	if len(encodedHex) != 64 {
		t.Errorf("Hex should be 32 bytes (64 hex lenght)")
	}

	secondEncodedHex, err := MakeRefreshToken()
	if err != nil {
		t.Fatalf("MakeRefreshToken function errored")
	}

	if encodedHex == secondEncodedHex {
		t.Errorf("The two hex are equivalent")
	}
}
