package acctest

import (
	"os"
	"testing"
)

var (
	TestApiKey    = os.Getenv("PORKBUN_API_KEY")
	TestSecretKey = os.Getenv("PORKBUN_SECRET_KEY")
)

func PreCheck(t *testing.T) {
	if TestApiKey == "" {
		t.Fatal("PORKBUN_API_KEY must be set for acceptance tests")
	}

	if TestSecretKey == "" {
		t.Fatal("PORKBUN_SECRET_KEY must be set for acceptance tests")
	}
}
