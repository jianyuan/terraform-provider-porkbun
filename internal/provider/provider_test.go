package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jianyuan/terraform-provider-porkbun/internal/acctest"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"porkbun": providerserver.NewProtocol6WithError(New("test")()),
}

func TestMain(m *testing.M) {
	if err := acctest.Setup(context.Background()); err != nil {
		panic(fmt.Sprintf("Error setting up acceptance tests: %s", err))
	}

	resource.TestMain(m)
}
