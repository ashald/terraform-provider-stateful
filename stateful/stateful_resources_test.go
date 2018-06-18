package stateful

import (
	"testing"

	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const template = `
resource "stateful_string" "object" { desired="%s" real="%s" }
resource "null_resource" "updates" { triggers { state="${stateful_string.object.hash}" } }
`

func getConfig(desired string, real string) string {
	return fmt.Sprintf(template, desired, real)
}

func TestStatefulString(t *testing.T) {
	var nullResourceId = new(string)

	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testProviders,
		Steps: []resource.TestStep{
			{
				Config:             getConfig("foo", ""), // initial
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					// hash should be derived from desired value
					testResourceAttrEquals("stateful_string.object", "hash", strPtr(getSHA256("foo"))),
					// Extract null resource's ID to track its recreation
					func(state *terraform.State) error {
						*nullResourceId = getResourceAttr(state, "null_resource.updates", "id")
						return nil
					},
				),
			},
			{
				Config:             getConfig("foo", "foo"), // do changes
				ExpectNonEmptyPlan: false,
				Check: resource.ComposeTestCheckFunc(
					// hash should be derived from desired value
					testResourceAttrEquals("stateful_string.object", "hash", strPtr(getSHA256("foo"))),
					// No diff -> null_resource should not get triggered
					testResourceAttrEquals("null_resource.updates", "id", nullResourceId),
				),
			},
			{
				Config:             getConfig("bar", "foo"), // desired value changed
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					// hash should be derived from desired value
					testResourceAttrEquals("stateful_string.object", "hash", strPtr(getSHA256("bar"))),
					// null_resource should be recreated
					testResourceAttrDoesNotEqual("null_resource.updates", "id", nullResourceId),
				),
			},
			{
				Config:             getConfig("bar", ""), // rogue real value
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					// hash should be derived from desired value
					testResourceAttrEquals("stateful_string.object", "hash", strPtr(getSHA256("bar"))),
					// null_resource should be recreated
					testResourceAttrDoesNotEqual("null_resource.updates", "id", nullResourceId),
				),
			},
		},
	})
}

func strPtr(t string) *string {
	return &t
}

func getResourceAttr(state *terraform.State, resource string, attr string) string {
	return state.RootModule().Resources[resource].Primary.Attributes[attr]
}

func testResourceAttrEquals(resource string, attr string, expected *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		realValue := getResourceAttr(state, resource, attr)
		if realValue != *expected {
			return fmt.Errorf("resource '%s' attribute '%s' value '%s' does not match expected '%s'", resource, attr, realValue, *expected)
		}
		return nil
	}
}

func testResourceAttrDoesNotEqual(resource string, attr string, expected *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		realValue := getResourceAttr(state, resource, attr)
		if realValue == *expected {
			return fmt.Errorf("resource '%s' attribute '%s' value '%s' should not equal '%s'", resource, attr, realValue, *expected)
		}
		return nil
	}
}
