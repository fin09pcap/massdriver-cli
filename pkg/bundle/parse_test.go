package bundle_test

import (
	"reflect"
	"testing"

	"github.com/massdriver-cloud/massdriver-cli/pkg/bundle"
)

func TestParseBundle(t *testing.T) {
	var got, _ = bundle.ParseBundle("./testdata/bundle.yaml")
	var want = bundle.Bundle{
		Schema:      "draft-07",
		Type:        "bundle",
		Name:        "aws-vpc",
		Description: "Something",
		Access:      "public",
		Artifacts:   map[string]interface{}{},
		Params: map[string]interface{}{
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":  "string",
					"title": "Name",
				},
				"age": map[string]interface{}{
					"type":  "integer",
					"title": "Age",
				},
			},
			"required": []interface{}{
				"name",
			},
		},
		Connections: map[string]interface{}{
			"required": []interface{}{
				"default",
			},
			"properties": map[string]interface{}{
				"default": map[string]interface{}{
					"type":  "string",
					"title": "Default credential",
				},
			},
		},
		Ui: map[string]interface{}{},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
