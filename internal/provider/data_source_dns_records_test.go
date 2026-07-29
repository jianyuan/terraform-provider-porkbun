package provider

import (
	"fmt"
	"regexp"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-porkbun/internal/acctest"
)

func TestAccDnsRecordsDataSource(t *testing.T) {
	rn := "data.porkbun_dns_records.test"
	content := sdkacctest.RandomWithPrefix("tf")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsRecordsDataSourceConfig(fmt.Sprintf(`
					type      = "TXT"
					content   = "%[1]s"
				`, content), ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.StringExact(acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("records"), knownvalue.SetPartial([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"type":    knownvalue.StringExact("TXT"),
							"content": knownvalue.StringExact(content),
						}),
					})),
				},
			},
		},
	})
}

func TestAccDnsRecordsDataSource_filterByType(t *testing.T) {
	rn := "data.porkbun_dns_records.test"
	content := sdkacctest.RandomWithPrefix("tf")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsRecordsDataSourceConfig(fmt.Sprintf(`
					type      = "TXT"
					content   = "%[1]s"
				`, content), `
					filter = {
						type = "TXT"
					}

					lifecycle {
						postcondition {
							condition     = length(self.records) > 0
							error_message = "expected at least one record"
						}

						postcondition {
							condition     = alltrue([for record in self.records : record.type == "TXT"])
							error_message = "expected all records to be of type TXT"
						}
					}
				`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.StringExact(acctest.TestDomain)),
				},
			},
		},
	})
}

func TestAccDnsRecordsDataSource_filterByTypeAndSubdomain(t *testing.T) {
	rn := "data.porkbun_dns_records.test"
	content := sdkacctest.RandomWithPrefix("tf")
	subdomain := sdkacctest.RandomWithPrefix("tf")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsRecordsDataSourceConfig(fmt.Sprintf(`
					type      = "TXT"
					content   = "%[1]s"
					subdomain = "%[2]s"
				`, content, subdomain), fmt.Sprintf(`
					filter = {
						type      = "TXT"
						subdomain = "%[1]s"
					}

					lifecycle {
						postcondition {
							condition     = length(self.records) > 0
							error_message = "expected at least one record"
						}

						postcondition {
							condition     = alltrue([for record in self.records : record.type == "TXT"])
							error_message = "expected all records to be of type TXT"
						}
					}
				`, subdomain)),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.StringExact(acctest.TestDomain)),
				},
			},
		},
	})
}

func TestAccDnsRecordsDataSource_validation(t *testing.T) {
	content := sdkacctest.RandomWithPrefix("tf")
	subdomain := sdkacctest.RandomWithPrefix("tf")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsRecordsDataSourceConfig(fmt.Sprintf(`
					type      = "TXT"
					content   = "%[1]s"
					subdomain = "%[2]s"
				`, content, subdomain), fmt.Sprintf(`
					filter = {
						subdomain = "%[1]s"
					}
				`, subdomain)),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta(`Attribute "filter.type" must be specified when "filter.subdomain" is`) + "\n" + regexp.QuoteMeta(`specified`)),
			},
		},
	})
}

func testAccDnsRecordsDataSourceConfig(recordExtras, extras string) string {
	return testAccDnsRecordResourceConfig(recordExtras) + fmt.Sprintf(`
data "porkbun_dns_records" "test" {
	domain = porkbun_dns_record.test.domain
	%[2]s
}
`, acctest.TestDomain, extras)
}
