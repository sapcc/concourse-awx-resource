package resource

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	fr "github.com/tbe/resource-framework/resource"
	"github.com/tbe/resource-framework/test"
)

func TestHandler_Out(t *testing.T) {
	input := "{ \"source\": { \"endpoint\": \"https://awx.global.cloud.sap\", \"auth_token\": \"" + os.Getenv("AWX_TOKEN") + "\" }, " +
		"\"params\": { \"template_id\": 917, \"inventory\": \"BB093\" } }"
	test.AutoTestOut(t, func() fr.Resource { return NewAWXResource() }, map[string]test.Case{
		"valid": {
			Input:  input,
			Output: `{ "version": {}, "metadata": [{}] }`,
			Validation: func(t *testing.T, assertions *assert.Assertions, res interface{}) {
				r := res.(*AWXResource)
				assert.Equal(t, "https://awx.global.cloud.sap", r.source.Endpoint)
			},
		},
	})
}
