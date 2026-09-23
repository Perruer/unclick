// Copyright 2026 Perruer
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package terraformutils

import (
	"errors"
	"strings"
	"testing"

	"github.com/zclconf/go-cty/cty"

	"github.com/Perruer/unclick/internal/schema"
)

func deploymentSchema() *schema.Block {
	labels := &schema.Attribute{Type: cty.Map(cty.String), Optional: true}
	return &schema.Block{
		Attributes: map[string]*schema.Attribute{
			"id":               {Type: cty.String, Optional: true, Computed: true},
			"wait_for_rollout": {Type: cty.Bool, Optional: true},
		},
		BlockTypes: map[string]*schema.NestedBlock{
			"metadata": {Nesting: schema.NestingList, MaxItems: 1, Block: schema.Block{
				Attributes: map[string]*schema.Attribute{
					"name":      {Type: cty.String, Optional: true},
					"namespace": {Type: cty.String, Optional: true},
					"labels":    labels,
				},
			}},
			"spec": {Nesting: schema.NestingList, MaxItems: 1, Block: schema.Block{
				Attributes: map[string]*schema.Attribute{
					"replicas": {Type: cty.String, Optional: true},
				},
				BlockTypes: map[string]*schema.NestedBlock{
					"selector": {Nesting: schema.NestingList, Block: schema.Block{
						Attributes: map[string]*schema.Attribute{"match_labels": labels},
					}},
				},
			}},
		},
	}
}

func TestWriteResourcesHCLKubernetesMaps(t *testing.T) {
	r := NewSimpleResource("unclick-e2e/web", "unclick-e2e/web", "kubernetes_deployment", "kubernetes", nil)
	r.Schema = deploymentSchema()
	r.Item = map[string]interface{}{
		"wait_for_rollout": "true",
		"metadata": []interface{}{map[string]interface{}{
			"name":      "web",
			"namespace": "unclick-e2e",
			"labels":    map[string]interface{}{"app": "web", "app.kubernetes.io/name": "web"},
		}},
		"spec": []interface{}{map[string]interface{}{
			"replicas": "1",
			"selector": []interface{}{map[string]interface{}{
				"match_labels": map[string]interface{}{"app": "web"},
			}},
		}},
	}
	out, err := WriteResourcesHCL([]Resource{r})
	if err != nil {
		t.Fatal(err)
	}
	got := squeeze(string(out))
	for _, want := range []string{
		`resource "kubernetes_deployment" "tfer--unclick-e2e-002F-web" {`,
		"wait_for_rollout = true",
		"match_labels = {",
		`"app.kubernetes.io/name" = "web"`,
		`replicas = "1"`,
		"selector {",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output lacks %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, `selector "match_labels"`) || strings.Contains(got, `metadata "labels"`) {
		t.Errorf("maps must not become labelled blocks:\n%s", got)
	}
}

func TestWriteResourcesHCLValues(t *testing.T) {
	s := &schema.Block{Attributes: map[string]*schema.Attribute{
		"vpc_id":  {Type: cty.String, Required: true},
		"name":    {Type: cty.String, Optional: true},
		"policy":  {Type: cty.String, Optional: true},
		"note":    {Type: cty.String, Optional: true},
		"port":    {Type: cty.Number, Optional: true},
		"ids":     {Type: cty.List(cty.String), Optional: true},
		"unknown": {Type: cty.String, Optional: true},
	}}
	r := NewSimpleResource("sg-1", "web", "aws_security_group", "aws", nil)
	r.Schema = s
	r.Item = map[string]interface{}{
		"vpc_id":        "${aws_vpc.tfer--main.id}",
		"name":          `say "hi" %{x}`,
		"policy":        "{\n  \"Resource\": \"arn:aws:s3:::b/$${aws:username}\"\n}\n",
		"note":          "prefix-${var.env}",
		"port":          "443",
		"ids":           []interface{}{"${aws_subnet.tfer--a.id}", "subnet-b"},
		"depends_on":    []string{"aws_vpc.tfer--main"},
		"not_in_schema": "dropped",
	}
	out, err := WriteResourcesHCL([]Resource{r})
	if err != nil {
		t.Fatal(err)
	}
	got := squeeze(string(out))
	for _, want := range []string{
		"vpc_id = aws_vpc.tfer--main.id",
		`name = "say \"hi\" %%{x}"`,
		"policy = <<EOT",
		"$${aws:username}",
		`note = "prefix-${var.env}"`,
		"port = 443",
		`ids = [aws_subnet.tfer--a.id, "subnet-b"]`,
		"depends_on = [aws_vpc.tfer--main]",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output lacks %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "not_in_schema") {
		t.Errorf("keys outside the schema must be dropped:\n%s", got)
	}
}

func TestWriteResourcesHCLNeedsSchema(t *testing.T) {
	r := NewSimpleResource("x", "x", "aws_vpc", "aws", nil)
	if _, err := WriteResourcesHCL([]Resource{r}); !errors.Is(err, ErrNoSchema) {
		t.Fatalf("err = %v, want ErrNoSchema", err)
	}
}

// squeeze collapses whitespace, so that checks do not depend on how
// hclwrite aligns the equals signs of neighbouring arguments.
func squeeze(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
