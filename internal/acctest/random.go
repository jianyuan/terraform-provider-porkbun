package acctest

import (
	"fmt"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

func RandomDomain() string {
	return fmt.Sprintf("%s.com", sdkacctest.RandomWithPrefix(domainPrefix))
}
