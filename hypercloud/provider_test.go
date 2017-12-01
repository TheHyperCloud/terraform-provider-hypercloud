package hypercloud

import (
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"hypercloud": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	u := os.Getenv("HC_BASE_URL")
	c := os.Getenv("HC_CREDENTIALS")
	if u != "" && c != "" {
		if !strings.HasPrefix(u, "https") {
			t.Fatalf("Base url %s is not https", u)
		}
	}
}
