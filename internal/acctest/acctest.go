package acctest

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
)

const (
	defaultBaseUrl         = "https://api.porkbun.com/api/json/v3"
	sandboxApiKeyPrefix    = "pk1_sb_"
	sandboxSecretKeyPrefix = "sk1_sb_"
	domainPrefix           = "tf"
)

var (
	TestBaseUrl   = os.Getenv("PORKBUN_BASE_URL")
	TestApiKey    = os.Getenv("PORKBUN_API_KEY")
	TestSecretKey = os.Getenv("PORKBUN_SECRET_KEY")
	TestDomain    string

	SharedClient *apiclient.ClientWithResponses
)

func init() {
	if TestBaseUrl == "" {
		TestBaseUrl = defaultBaseUrl
	}

	if TestApiKey == "" || !strings.HasPrefix(TestApiKey, sandboxApiKeyPrefix) {
		panic("PORKBUN_API_KEY must start with 'pk1_sb_' to use the sandbox API")
	}

	if TestSecretKey == "" || !strings.HasPrefix(TestSecretKey, sandboxSecretKeyPrefix) {
		panic("PORKBUN_SECRET_KEY must start with 'sk1_sb_' to use the sandbox API")
	}
}

func PreCheck(t *testing.T) {
}

func Setup(ctx context.Context) error {
	client, err := apiclient.New(TestBaseUrl, TestApiKey, TestSecretKey)
	if err != nil {
		return fmt.Errorf("failed to create Porkbun client: %w", err)
	}
	SharedClient = client

	if os.Getenv("PORKBUN_RESET_SANDBOX") == "true" {
		if err := resetSandbox(ctx); err != nil {
			return fmt.Errorf("failed to reset sandbox: %w", err)
		}
	}

	if err := ensureTestDomain(ctx); err != nil {
		return fmt.Errorf("failed to ensure test domain: %w", err)
	}

	return nil
}

func resetSandbox(ctx context.Context) error {
	resetHttpResp, err := SharedClient.SandboxResetWithResponse(ctx, apiclient.SandboxResetJSONRequestBody{})
	if err != nil {
		return fmt.Errorf("failed to reset sandbox: %w", err)
	} else if resetHttpResp.StatusCode() != http.StatusOK || resetHttpResp.JSON200 == nil {
		return fmt.Errorf("failed to reset sandbox, status: %d, body: %v", resetHttpResp.StatusCode(), string(resetHttpResp.Body))
	}

	topupHttpResp, err := SharedClient.SandboxTopupWithResponse(ctx, apiclient.SandboxTopupJSONRequestBody{
		Amount: new(int64(1000000)),
	})
	if err != nil {
		return fmt.Errorf("failed to topup sandbox: %w", err)
	} else if topupHttpResp.StatusCode() != http.StatusOK || topupHttpResp.JSON200 == nil {
		return fmt.Errorf("failed to topup sandbox, status: %d, body: %v", topupHttpResp.StatusCode(), string(topupHttpResp.Body))
	}

	return nil
}

func ensureTestDomain(ctx context.Context) error {
	if TestDomain != "" {
		return nil
	}

	searchPrefix := fmt.Sprintf("%s-", domainPrefix)
	listDomainsHttpResp, err := SharedClient.ListDomainsWithResponse(ctx, apiclient.ListDomainsJSONRequestBody{
		NameContains: &searchPrefix,
	})
	if err != nil {
		return fmt.Errorf("failed to list domains: %w", err)
	} else if listDomainsHttpResp.StatusCode() != http.StatusOK || listDomainsHttpResp.JSON200 == nil {
		return fmt.Errorf("failed to list domains, status: %d, body: %v", listDomainsHttpResp.StatusCode(), string(listDomainsHttpResp.Body))
	}

	for _, domain := range listDomainsHttpResp.JSON200.Domains {
		if domain.Domain != nil && strings.HasPrefix(*domain.Domain, searchPrefix) {
			TestDomain = *domain.Domain
			return nil
		}
	}

	newDomain := fmt.Sprintf("%s.com", sdkacctest.RandomWithPrefix(domainPrefix))
	domainCheckHttpResp, err := SharedClient.DomainCheckDomainWithResponse(ctx, newDomain, apiclient.DomainCheckDomainJSONRequestBody{})
	if err != nil {
		return fmt.Errorf("failed to check domain availability: %w", err)
	} else if domainCheckHttpResp.StatusCode() != http.StatusOK || domainCheckHttpResp.JSON200 == nil || domainCheckHttpResp.JSON200.Status != "SUCCESS" {
		return fmt.Errorf("failed to check domain availability, status: %d, body: %v", domainCheckHttpResp.StatusCode(), string(domainCheckHttpResp.Body))
	} else if domainCheckHttpResp.JSON200.Response.Avail == nil || *domainCheckHttpResp.JSON200.Response.Avail != "yes" {
		return fmt.Errorf("domain %s is not available", newDomain)
	} else if domainCheckHttpResp.JSON200.Response.Price == nil {
		return fmt.Errorf("domain %s has no price", newDomain)
	}

	price, err := strconv.ParseFloat(*domainCheckHttpResp.JSON200.Response.Price, 64)
	if err != nil {
		return fmt.Errorf("failed to parse price: %w", err)
	}

	domainCreateHttpResp, err := SharedClient.DomainCreateWithResponse(ctx, newDomain, apiclient.DomainCreateJSONRequestBody{
		AgreeToTerms: "yes",
		Cost:         int64(math.Round(price * 100)),
	})
	if err != nil {
		return fmt.Errorf("failed to create domain: %w", err)
	} else if domainCreateHttpResp.StatusCode() != http.StatusOK || domainCreateHttpResp.JSON200 == nil {
		return fmt.Errorf("failed to create domain, status: %d, body: %v", domainCreateHttpResp.StatusCode(), string(domainCreateHttpResp.Body))
	}

	domainCreateResp, err := domainCreateHttpResp.JSON200.AsCreateDomainResponse()
	if err != nil {
		return fmt.Errorf("failed to parse create domain response: %w", err)
	} else if domainCreateResp.Status != "SUCCESS" {
		return fmt.Errorf("failed to create domain, status: %s, response: %v", domainCreateResp.Status, domainCreateResp)
	}

	TestDomain = newDomain
	return nil
}
