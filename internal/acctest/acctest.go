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

var (
	TestBaseUrl   = os.Getenv("PORKBUN_BASE_URL")
	TestApiKey    = os.Getenv("PORKBUN_API_KEY")
	TestSecretKey = os.Getenv("PORKBUN_SECRET_KEY")
	TestDomain    = ""

	SharedClient *apiclient.ClientWithResponses
)

func init() {
	if TestBaseUrl == "" {
		TestBaseUrl = "https://api.porkbun.com/api/json/v3"
	}

	if TestApiKey == "" || !strings.HasPrefix(TestApiKey, "pk1_sb_") {
		panic("PORKBUN_API_KEY must start with 'pk1_sb_' to use the sandbox API")
	}

	if TestSecretKey == "" || !strings.HasPrefix(TestSecretKey, "sk1_sb_") {
		panic("PORKBUN_SECRET_KEY must start with 'sk1_sb_' to use the sandbox API")
	}
}

func PreCheck(t *testing.T) {
}

func Setup(ctx context.Context) error {
	var err error
	SharedClient, err = apiclient.New(TestBaseUrl, TestApiKey, TestSecretKey)
	if err != nil {
		return fmt.Errorf("failed to create Porkbun client: %w", err)
	}

	if os.Getenv("PORKBUN_RESET_SANDBOX") == "true" {
		err = resetSandbox(ctx)
		if err != nil {
			return fmt.Errorf("failed to reset sandbox: %w", err)
		}
	}

	err = ensureTestDomain(ctx)
	if err != nil {
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

	topupHttpResp, err := SharedClient.SandboxTopupWithResponse(ctx, apiclient.SandboxTopupJSONRequestBody{Amount: new(int64(1000000))})
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

	TestDomain = sdkacctest.RandomWithPrefix("tf") + ".com"
	domainCheckHttpResp, err := SharedClient.DomainCheckDomainWithResponse(ctx, TestDomain, apiclient.DomainCheckDomainJSONRequestBody{})
	if err != nil {
		return fmt.Errorf("failed to check domain availability: %w", err)
	} else if domainCheckHttpResp.StatusCode() != http.StatusOK || domainCheckHttpResp.JSON200 == nil || domainCheckHttpResp.JSON200.Status != "SUCCESS" {
		return fmt.Errorf("failed to check domain availability, status: %d, body: %v", domainCheckHttpResp.StatusCode(), string(domainCheckHttpResp.Body))
	} else if domainCheckHttpResp.JSON200.Response.Avail == nil || *domainCheckHttpResp.JSON200.Response.Avail != "yes" {
		return fmt.Errorf("domain %s is not available", TestDomain)
	}

	price, err := strconv.ParseFloat(*domainCheckHttpResp.JSON200.Response.Price, 64)
	if err != nil {
		return fmt.Errorf("failed to parse price: %w", err)
	}

	domainCreateHttpResp, err := SharedClient.DomainCreateWithResponse(ctx, TestDomain, apiclient.DomainCreateJSONRequestBody{
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

	return nil
}
