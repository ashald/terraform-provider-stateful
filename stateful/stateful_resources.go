package stateful

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/satori/go.uuid"
	"reflect"
)

const FieldDesired = "desired"
const FieldReal = "real"

const FieldHash = "hash"

func resourceStatefulString() *schema.Resource {
	return resourceFactory(schema.TypeString)
}

func resourceStatefulMap() *schema.Resource {
	return resourceFactory(schema.TypeMap)
}

func resourceFactory(inputType schema.ValueType) *schema.Resource {
	return &schema.Resource{
		Create: createResource,
		Read:   readResource,
		Update: updateResource,
		Delete: deleteResource,

		CustomizeDiff: diffResource,

		Schema: map[string]*schema.Schema{
			// "Inputs"
			FieldDesired: {
				Type:     inputType,
				Required: true,
			},
			FieldReal: {
				Type:     inputType,
				Optional: true,
				Computed: true,
			},
			// "Outputs"
			FieldHash: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func getSHA256(o interface{}) string {
	serialized, _ := json.Marshal(o)
	h := sha256.New()
	h.Write([]byte(serialized))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func getStatefulResourceFingerprint(d *schema.ResourceData) string {
	data := d.Get(FieldDesired)
	return getSHA256(data)
}

func createResource(d *schema.ResourceData, m interface{}) error {
	d.SetId(uuid.Must(uuid.NewV4()).String())

	sha256hash := getStatefulResourceFingerprint(d)
	d.Set(FieldHash, sha256hash)

	return nil
}

func readResource(d *schema.ResourceData, m interface{}) error {
	sha256hash := getStatefulResourceFingerprint(d)
	d.Set(FieldHash, sha256hash)
	return nil
}

func updateResource(d *schema.ResourceData, m interface{}) error {
	sha256hash := getStatefulResourceFingerprint(d)
	d.Set(FieldHash, sha256hash)
	return nil
}

func deleteResource(d *schema.ResourceData, m interface{}) error {
	return nil
}

func diffResource(d *schema.ResourceDiff, m interface{}) error {
	desiredValue := d.Get(FieldDesired)
	realValue, realValueIsSet := d.GetOkExists(FieldReal)

	if realValueIsSet {
		if reflect.DeepEqual(desiredValue, realValue) {
			d.Clear(FieldReal)
		} else {
			d.SetNewComputed(FieldReal)
			d.SetNewComputed(FieldHash)
		}
	} else {
		d.Clear(FieldReal)
	}

	if d.HasChange(FieldDesired) {
		d.SetNewComputed(FieldHash)
	}

	return nil
}
