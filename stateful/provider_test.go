package stateful

import (
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-null/null"
)

var testProviders map[string]terraform.ResourceProvider

var nullProvider *schema.Provider
var statefulProvider *schema.Provider

func init() {
	statefulProvider = Provider().(*schema.Provider)
	nullProvider = null.Provider().(*schema.Provider)

	testProviders = map[string]terraform.ResourceProvider{
		"stateful": statefulProvider,
		"null":     nullProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := statefulProvider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
