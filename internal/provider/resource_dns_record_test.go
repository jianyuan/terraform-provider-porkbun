package provider

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-porkbun/internal/acctest"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
)

func init() {
	resource.AddTestSweepers("porkbun_dns_record", &resource.Sweeper{
		Name: "porkbun_dns_record",
		F: func(r string) error {
			ctx := context.Background()

			httpResp, err := acctest.SharedClient.DnsRetrieveRecordsByDomainWithResponse(
				ctx,
				acctest.TestDomain,
				apiclient.DnsRetrieveRecordsByDomainJSONRequestBody{
					Apikey:       acctest.TestApiKey,
					Secretapikey: acctest.TestSecretKey,
				},
			)
			if err != nil {
				return fmt.Errorf("Unable to read, got error: %s", err)
			} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil || httpResp.JSON200.Status != "SUCCESS" {
				return fmt.Errorf("Unable to read, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body))
			}

			for _, record := range httpResp.JSON200.Records {
				if !strings.HasPrefix(record.Name, "tf-") && !strings.HasPrefix(record.Content, "tf-") {
					continue
				}

				log.Printf("[INFO] Destroying record %s", record.Id)

				_, err := acctest.SharedClient.DnsDeleteRecordByDomainAndIdWithResponse(
					ctx,
					acctest.TestDomain,
					record.Id,
					apiclient.DnsDeleteRecordByDomainAndIdJSONRequestBody{
						Apikey:       acctest.TestApiKey,
						Secretapikey: acctest.TestSecretKey,
					},
				)

				if err != nil {
					log.Printf("[ERROR] Unable to delete record %s: %s", record.Id, err)
					continue
				}

				log.Printf("[INFO] Deleted record %s", record.Id)
			}

			return nil
		},
	})
}

func TestAccDnsRecordResource(t *testing.T) {
	rn := "porkbun_dns_record.test"
	content := acctest.RandomWithPrefix("tf")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsRecordResourceConfig(fmt.Sprintf(`
					type    = "TXT"
					content = "%[1]s"
				`, content)),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("subdomain"), knownvalue.Null()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.StringExact(acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("type"), knownvalue.StringExact("TXT")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("content"), knownvalue.StringExact(content)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("ttl"), knownvalue.Int64Exact(600)),
				},
			},
			{
				Config: testAccDnsRecordResourceConfig(fmt.Sprintf(`
					subdomain = "%[1]s"
					type      = "TXT"
					content   = "%[1]s"
				`, content)),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("subdomain"), knownvalue.StringExact(content)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.StringExact(content+"."+acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("type"), knownvalue.StringExact("TXT")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("content"), knownvalue.StringExact(content)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("ttl"), knownvalue.Int64Exact(600)),
				},
			},
			{
				Config: testAccDnsRecordResourceConfig(fmt.Sprintf(`
					type    = "TXT"
					content = "%[1]s"
					ttl     = 300
				`, content)),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("subdomain"), knownvalue.Null()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.StringExact(acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("type"), knownvalue.StringExact("TXT")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("content"), knownvalue.StringExact(content)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("ttl"), knownvalue.Int64Exact(300)),
				},
			},
			{
				Config: testAccDnsRecordResourceConfig(fmt.Sprintf(`
					subdomain = "*"
					type      = "TXT"
					content   = "%[1]s"
					ttl       = 300
				`, content)),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("subdomain"), knownvalue.StringExact("*")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.StringExact("*."+acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("type"), knownvalue.StringExact("TXT")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("content"), knownvalue.StringExact(content)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("ttl"), knownvalue.Int64Exact(300)),
				},
			},
		},
	})
}

func testAccDnsRecordResourceConfig(extras string) string {
	return fmt.Sprintf(`
resource "porkbun_dns_record" "test" {
	domain = "%[1]s"
	%[2]s
}
`, acctest.TestDomain, extras)
}
