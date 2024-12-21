package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-porkbun/internal/acctest"
)

func TestAccDnsRecordsDataSource(t *testing.T) {
	rn := "data.porkbun_dns_records.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsRecordsDataSourceConfig(""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.StringExact(acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("records"), knownvalue.SetPartial([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"name": knownvalue.StringExact(acctest.TestDomain),
							"type": knownvalue.StringExact("ALIAS"),
						}),
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"name": knownvalue.StringExact("*." + acctest.TestDomain),
							"type": knownvalue.StringExact("CNAME"),
						}),
					})),
				},
			},
		},
	})
}

func TestAccDnsRecordsDataSource_filterByType(t *testing.T) {
	rn := "data.porkbun_dns_records.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsRecordsDataSourceConfig(`
					filter = {
						type = "ALIAS"
					}

					lifecycle {
						postcondition {
							condition     = length(self.records) > 0
							error_message = "expected at least one record"
						}

						postcondition {
							condition     = alltrue([for record in self.records : record.type == "ALIAS"])
							error_message = "expected all records to be of type ALIAS"
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

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsRecordsDataSourceConfig(`
					filter = {
						type      = "A"
						subdomain = "www"
					}

					lifecycle {
						postcondition {
							condition     = length(self.records) > 0
							error_message = "expected at least one record"
						}

						postcondition {
							condition     = alltrue([for record in self.records : record.type == "A"])
							error_message = "expected all records to be of type A"
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

func TestAccDnsRecordsDataSource_validation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsRecordsDataSourceConfig(`
					filter = {
						subdomain = "www"
					}
				`),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta(`Attribute "filter.type" must be specified when "filter.subdomain" is`) + "\n" + regexp.QuoteMeta(`specified`)),
			},
		},
	})
}

func testAccDnsRecordsDataSourceConfig(extras string) string {
	return fmt.Sprintf(`
data "porkbun_dns_records" "test" {
	domain = "%[1]s"
	%[2]s
}
`, acctest.TestDomain, extras)
}
