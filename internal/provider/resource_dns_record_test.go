package provider

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-porkbun/internal/acctest"
)

func TestAccDnsRecordResource(t *testing.T) {
	rn := "porkbun_dns_record.test"
	content := sdkacctest.RandomWithPrefix("tf")

	subdomainConfig := testAccDnsRecordResourceConfig(fmt.Sprintf(`
		subdomain = "%[1]s"
		type      = "TXT"
		content   = "%[1]s"
	`, content))
	wildcardConfig := testAccDnsRecordResourceConfig(fmt.Sprintf(`
		subdomain = "*"
		type      = "TXT"
		content   = "%[1]s"
		ttl       = 300
	`, content))

	emptyPlan := resource.ConfigPlanChecks{
		PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
	}

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
				Config: subdomainConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("subdomain"), knownvalue.StringExact(content)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.StringExact(content+"."+acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("type"), knownvalue.StringExact("TXT")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("content"), knownvalue.StringExact(content)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("ttl"), knownvalue.Int64Exact(600)),
				},
			},
			{
				Config:            subdomainConfig,
				ResourceName:      rn,
				ImportState:       true,
				ImportStateIdFunc: testAccDnsRecordImportStateIdFunc(rn),
				ImportStateVerify: true,
			},
			{
				Config:           subdomainConfig,
				ConfigPlanChecks: emptyPlan,
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
				Config: wildcardConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("subdomain"), knownvalue.StringExact("*")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.StringExact("*."+acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("type"), knownvalue.StringExact("TXT")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("content"), knownvalue.StringExact(content)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("ttl"), knownvalue.Int64Exact(300)),
				},
			},
			{
				Config:            wildcardConfig,
				ResourceName:      rn,
				ImportState:       true,
				ImportStateIdFunc: testAccDnsRecordImportStateIdFunc(rn),
				ImportStateVerify: true,
			},
			{
				Config:           wildcardConfig,
				ConfigPlanChecks: emptyPlan,
			},
		},
	})
}

func testAccDnsRecordImportStateIdFunc(rn string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return "", fmt.Errorf("not found: %s", rn)
		}
		return fmt.Sprintf("%s_%s_%s", rs.Primary.ID, rs.Primary.Attributes["domain"], rs.Primary.Attributes["type"]), nil
	}
}

func testAccDnsRecordResourceConfig(extras string) string {
	return fmt.Sprintf(`
resource "porkbun_dns_record" "test" {
	domain = "%[1]s"
	%[2]s
}
`, acctest.TestDomain, extras)
}
