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
	"fmt"
	"strings"
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
	if f.calls != 3 {
		t.Errorf("validated %d times, want 3 (one fix per round, then give up on the required one)", f.calls)
	}
	if len(r.LeftOut) != 2 {
		t.Errorf("LeftOut = %q, want the two removed arguments", r.LeftOut)
	}
	if len(r.ProviderErrors) != 1 || r.ProviderErrors[0] != "vpc_id: bad vpc" {
		t.Errorf("ProviderErrors = %q", r.ProviderErrors)
	}
}

// conflictValidator reports cidr_block and outpost_arn as conflicting while
// both are set, the way ConflictsWith does.
type conflictValidator struct{ s *schema.Provider }

func (c conflictValidator) GetSchema() *schema.Provider { return c.s }

func (c conflictValidator) ValidateResourceConfig(_ string, v cty.Value) ([]plugin.Diagnostic, error) {
	if v.GetAttr("cidr_block").IsNull() || v.GetAttr("outpost_arn").IsNull() {
		return nil, nil
	}
	return []plugin.Diagnostic{
		{Error: true, Summary: "conflicts with outpost_arn", Path: []plugin.PathStep{{Attribute: "cidr_block"}}},
		{Error: true, Summary: "conflicts with cidr_block", Path: []plugin.PathStep{{Attribute: "outpost_arn"}}},
	}, nil
}

func TestFixInvalidConfigDropsTheZeroValueFirst(t *testing.T) {
	r := NewSimpleResource("subnet-1", "a", "aws_subnet", "aws", nil)
	r.Item = map[string]interface{}{"vpc_id": "vpc-1", "cidr_block": "10.0.0.0/24", "outpost_arn": ""}
	invalid := FixInvalidConfig([]*Resource{&r}, conflictValidator{subnetSchema()})

	if len(invalid) != 0 {
		t.Fatalf("still invalid: %q", r.ProviderErrors)
	}
	if r.Item["cidr_block"] != "10.0.0.0/24" {
		t.Error("the argument with a real value was removed instead of the empty one")
	}
	if _, ok := r.Item["outpost_arn"]; ok {
		t.Error("the empty conflicting argument should be gone")
	}
}

// chainValidator rejects the lowest-numbered argument left, so every fix
// uncovers the next conflict: a chain longer than any fixed round limit.
type chainValidator struct {
	s     *schema.Provider
	calls int
}

func (c *chainValidator) GetSchema() *schema.Provider { return c.s }

func (c *chainValidator) ValidateResourceConfig(_ string, v cty.Value) ([]plugin.Diagnostic, error) {
	c.calls++
	for i := 0; i < 20; i++ {
		name := fmt.Sprintf("a%02d", i)
		if !v.GetAttr(name).IsNull() {
			return []plugin.Diagnostic{{Error: true, Summary: "conflicts with the next one", Path: []plugin.PathStep{{Attribute: name}}}}, nil
		}
	}
	return nil, nil
}

func TestFixInvalidConfigFollowsLongChains(t *testing.T) {
	attrs := map[string]*schema.Attribute{"name": {Type: cty.String, Required: true}}
	item := map[string]interface{}{"name": "x"}
	for i := 0; i < 20; i++ {
		name := fmt.Sprintf("a%02d", i)
		attrs[name] = &schema.Attribute{Type: cty.String, Optional: true}
		item[name] = "v"
	}
	s := &schema.Provider{ResourceTypes: map[string]schema.Resource{"x_thing": {Block: &schema.Block{Attributes: attrs}}}}
	r := NewSimpleResource("x-1", "x", "x_thing", "x", nil)
	r.Item = item
	v := &chainValidator{s: s}

	if invalid := FixInvalidConfig([]*Resource{&r}, v); len(invalid) != 0 {
		t.Fatalf("still invalid after the chain: %q", r.ProviderErrors)
	}
	if len(r.LeftOut) != 20 || v.calls != 21 {
		t.Errorf("left out %d arguments in %d validations, want 20 in 21", len(r.LeftOut), v.calls)
	}
}

func TestWriteResourcesHCLExplainsFixes(t *testing.T) {
	r := NewSimpleResource("subnet-1", "a", "aws_subnet", "aws", nil)
	r.Item = map[string]interface{}{"vpc_id": "bad"}
	r.Schema = subnetSchema().ResourceTypes["aws_subnet"].Block
	r.LeftOut = []string{"outpost_arn (conflicts with cidr_block)"}
	r.ProviderErrors = []string{"vpc_id: bad vpc:\nsee the docs"}

	out, err := WriteResourcesHCL([]Resource{r})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"# unclick: left out outpost_arn (conflicts with cidr_block)",
		"# unclick: the provider still rejects this resource",
		"#   vpc_id: bad vpc: see the docs",
	} {
		if !strings.Contains(string(out), want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
}
