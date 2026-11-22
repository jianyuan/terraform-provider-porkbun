package provider

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/jianyuan/terraform-provider-porkbun/internal/acctest"
)

func TestAccEmailForwardsDataSource(t *testing.T) {
	t.Parallel()

	if !acctest.TestRunEmailForward {
		t.Skip("PORKBUN_RUN_EMAIL_FORWARD is not set to 1; the sandbox does not support /email endpoints")
	}

	rn := "data.porkbun_email_forwards.test"
	address := sdkacctest.RandomWithPrefix("tf")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEmailForwardResourceConfig(address) + `
data "porkbun_email_forwards" "test" {
	domain = porkbun_email_forward.test.domain
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("domain"), knownvalue.StringExact(acctest.TestDomain)),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("dns_configured"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("limits").AtMapKey("max"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(rn, tfjsonpath.New("email_forwards"), knownvalue.SetPartial([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"address":    knownvalue.StringExact(address),
							"forward_to": knownvalue.StringExact("me@example.org"),
							"created":    knownvalue.NotNull(),
						}),
					})),
				},
			},
		},
	})
}
