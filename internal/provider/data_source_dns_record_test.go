package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-porkbun/internal/acctest"
)

func TestAccDnsRecordDataSource(t *testing.T) {
	rn := "data.porkbun_dns_record.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsRecordDataSourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.StringExact(acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("type"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("content"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func testAccDnsRecordDataSourceConfig() string {
	return fmt.Sprintf(`
data "porkbun_dns_records" "test" {
	domain = "%[1]s"

	lifecycle {
		postcondition {
			condition     = length(self.records) > 0
			error_message = "expected at least one record"
		}
	}
}

locals {
	first_record = tolist(data.porkbun_dns_records.test.records)[0]
}

data "porkbun_dns_record" "test" {
	domain = data.porkbun_dns_records.test.domain
	id     = local.first_record.id
}
`, acctest.TestDomain)
}
