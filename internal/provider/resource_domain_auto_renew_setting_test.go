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

func TestAccDomainAutoRenewSettingResource(t *testing.T) {
	rn := "porkbun_domain_auto_renew_setting.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainAutoRenewSettingResourceConfig(true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("auto_renew"), knownvalue.Bool(true)),
				},
			},
			{
				Config: testAccDomainAutoRenewSettingResourceConfig(false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("auto_renew"), knownvalue.Bool(false)),
				},
			},
			{
				Config:                               testAccDomainAutoRenewSettingResourceConfig(false),
				ResourceName:                         rn,
				ImportState:                          true,
				ImportStateId:                        acctest.TestDomain,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "domain",
			},
		},
	})
}

func testAccDomainAutoRenewSettingResourceConfig(autoRenew bool) string {
	return fmt.Sprintf(`
resource "porkbun_domain_auto_renew_setting" "test" {
	domain     = "%s"
	auto_renew = %t
}`,
		acctest.TestDomain, autoRenew)
}
