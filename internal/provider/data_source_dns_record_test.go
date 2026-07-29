package provider

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-porkbun/internal/acctest"
)

func TestAccDnsRecordDataSource(t *testing.T) {
	rn := "data.porkbun_dns_record.test"
	content := sdkacctest.RandomWithPrefix("tf")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsRecordDataSourceConfig(content),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.StringExact(acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("name"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("type"), knownvalue.StringExact("TXT")),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("content"), knownvalue.StringExact(content)),
				},
			},
		},
	})
}

func testAccDnsRecordDataSourceConfig(content string) string {
	return testAccDnsRecordResourceConfig(fmt.Sprintf(`
type    = "TXT"
content = "%[1]s"
`, content)) + `
data "porkbun_dns_record" "test" {
	domain = porkbun_dns_record.test.domain
	id     = porkbun_dns_record.test.id
}
`
}
