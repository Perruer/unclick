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
	"testing"

	"github.com/zclconf/go-cty/cty"

	"github.com/Perruer/unclick/internal/plugin"
	"github.com/Perruer/unclick/internal/schema"
)

func subnetSchema() *schema.Provider {
	return &schema.Provider{ResourceTypes: map[string]schema.Resource{
		"aws_subnet": {Block: &schema.Block{
			Attributes: map[string]*schema.Attribute{
				"id":                              {Type: cty.String, Optional: true, Computed: true},
				"vpc_id":                          {Type: cty.String, Required: true},
				"cidr_block":                      {Type: cty.String, Optional: true},
				"map_customer_owned_ip_on_launch": {Type: cty.Bool, Optional: true},
				"outpost_arn":                     {Type: cty.String, Optional: true},
				"tags":                            {Type: cty.Map(cty.String), Optional: true},
				"ports":                           {Type: cty.Set(cty.Number), Optional: true},
			},
			BlockTypes: map[string]*schema.NestedBlock{
				"timeouts": {Nesting: schema.NestingSingle, Block: schema.Block{
					Attributes: map[string]*schema.Attribute{"create": {Type: cty.String, Optional: true}},
				}},
				"rule": {Nesting: schema.NestingList, Block: schema.Block{
					Attributes: map[string]*schema.Attribute{
						"port":        {Type: cty.Number, Required: true},
						"description": {Type: cty.String, Optional: true},
					},
				}},
			},
		}},
	}}
}

func TestItemValue(t *testing.T) {
	b := subnetSchema().ResourceTypes["aws_subnet"].Block
	v, err := ItemValue(map[string]interface{}{
		"vpc_id":                          "${aws_vpc.tfer--main.id}",
		"cidr_block":                      "10.0.1.0/24",
		"map_customer_owned_ip_on_launch": "false",
		"tags":                            map[string]interface{}{"Name": "a"},
		"ports":                           []interface{}{"443", "80"},
		"rule":                            []interface{}{map[string]interface{}{"port": "22"}},
	}, b)
	if err != nil {
		t.Fatal(err)
	}
	if errs := v.Type().TestConformance(b.ImpliedType()); len(errs) > 0 {
		t.Fatalf("value does not conform: %v", errs)
	}
	if v.GetAttr("vpc_id").IsKnown() {
		t.Error("a reference must become an unknown value")
	}
	if !v.GetAttr("map_customer_owned_ip_on_launch").RawEquals(cty.False) {
		t.Errorf("bool = %#v", v.GetAttr("map_customer_owned_ip_on_launch"))
	}
	if !v.GetAttr("outpost_arn").IsNull() {
		t.Error("missing attribute must be null")
	}
	if !v.GetAttr("timeouts").IsNull() {
		t.Error("missing single block must be null")
	}
	port := v.GetAttr("rule").Index(cty.NumberIntVal(0)).GetAttr("port")
	if !port.RawEquals(cty.NumberIntVal(22)) {
		t.Errorf("rule[0].port = %#v", port)
	}
}

// fakeValidator mimics the AWS provider: map_customer_owned_ip_on_launch
// requires outpost_arn, and rule descriptions must not be empty.
type fakeValidator struct {
	s     *schema.Provider
	calls int
}

func (f *fakeValidator) GetSchema() *schema.Provider { return f.s }

func (f *fakeValidator) ValidateResourceConfig(_ string, v cty.Value) ([]plugin.Diagnostic, error) {
	f.calls++
	var diags []plugin.Diagnostic
	if !v.GetAttr("map_customer_owned_ip_on_launch").IsNull() && v.GetAttr("outpost_arn").IsNull() {
		diags = append(diags, plugin.Diagnostic{
			Error:   true,
			Summary: "Missing required argument",
			Path:    []plugin.PathStep{{Attribute: "map_customer_owned_ip_on_launch"}},
		})
	}
	for it := v.GetAttr("rule").ElementIterator(); it.Next(); {
		k, rule := it.Element()
		if d := rule.GetAttr("description"); !d.IsNull() && d.AsString() == "" {
			i, _ := k.AsBigFloat().Int64()
			diags = append(diags, plugin.Diagnostic{
				Error:   true,
				Summary: "expected description to not be an empty string",
				Path:    []plugin.PathStep{{Attribute: "rule"}, {Index: i, IsIndex: true}, {Attribute: "description"}},
			})
		}
	}
	// A rule on a required argument cannot be fixed by leaving it out.
	if v.GetAttr("vpc_id").IsKnown() && v.GetAttr("vpc_id").AsString() == "bad" {
		diags = append(diags, plugin.Diagnostic{Error: true, Summary: "bad vpc", Path: []plugin.PathStep{{Attribute: "vpc_id"}}})
	}
	return diags, nil
}

func TestFixInvalidConfig(t *testing.T) {
	r := NewSimpleResource("subnet-1", "a", "aws_subnet", "aws", nil)
	r.Item = map[string]interface{}{
		"vpc_id":                          "bad",
		"map_customer_owned_ip_on_launch": "false",
		"rule": []interface{}{
			map[string]interface{}{"port": "22", "description": "ssh"},
			map[string]interface{}{"port": "80", "description": ""},
		},
	}
	f := &fakeValidator{s: subnetSchema()}
	FixInvalidConfig([]*Resource{&r}, f)

	if _, ok := r.Item["map_customer_owned_ip_on_launch"]; ok {
		t.Error("the rejected optional argument should be gone")
	}
	if r.Item["vpc_id"] != "bad" {
		t.Error("a required argument must never be removed")
	}
	rules := r.Item["rule"].([]interface{})
	if _, ok := rules[1].(map[string]interface{})["description"]; ok {
		t.Error("the rejected nested argument should be gone")
	}
	if rules[0].(map[string]interface{})["description"] != "ssh" {
		t.Error("only the element the diagnostic points at may change")
	}
	if f.calls != 2 {
		t.Errorf("validated %d times, want 2 (fix, then give up on the required one)", f.calls)
	}
}
