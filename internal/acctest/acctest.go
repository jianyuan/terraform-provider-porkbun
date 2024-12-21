package acctest

import (
	"os"
	"testing"

	"github.com/jianyuan/go-utils/must"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
)

var (
	TestApiKey    = os.Getenv("PORKBUN_API_KEY")
	TestSecretKey = os.Getenv("PORKBUN_SECRET_KEY")
	TestDomain    = os.Getenv("PORKBUN_TEST_DOMAIN")

	SharedClient *apiclient.ClientWithResponses
)

func init() {
	SharedClient = must.Get(apiclient.NewClientWithResponses(
		"https://api.porkbun.com/api/json",
	))
}

func PreCheck(t *testing.T) {
	if TestApiKey == "" {
		t.Fatal("PORKBUN_API_KEY must be set for acceptance tests")
	}

	if TestSecretKey == "" {
		t.Fatal("PORKBUN_SECRET_KEY must be set for acceptance tests")
	}

	if TestDomain == "" {
		t.Fatal("PORKBUN_TEST_DOMAIN must be set for acceptance tests")
	}
}
