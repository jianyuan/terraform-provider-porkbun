package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-porkbun/internal/acctest"
)

func TestAccAPISpendSettingsDataSource(t *testing.T) {
	t.Parallel()

	rn := "data.porkbun_api_spend_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAPISpendSettingsDataSourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("monthly_spend"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("settings"), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"auto_topup":          knownvalue.Bool(false),
						"low_balance_alert":   knownvalue.Null(),
						"monthly_spend_limit": knownvalue.Null(),
						"topup_amount":        knownvalue.Null(),
						"topup_threshold":     knownvalue.Null(),
					})),
				},
			},
		},
	})
}

func testAccAPISpendSettingsDataSourceConfig() string {
	return `
data "porkbun_api_spend_settings" "test" {
}
`
}
