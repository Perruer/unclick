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

package schema

import (
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func testBlock() *Block {
	return &Block{
		Attributes: map[string]*Attribute{
			"name":    {Type: cty.String, Required: true},
			"count":   {Type: cty.Number, Optional: true},
			"arn":     {Type: cty.String, Computed: true},
			"dynamic": {Type: cty.DynamicPseudoType, Optional: true},
			"rules": {NestedType: &Object{
				Nesting:    NestingList,
				Attributes: map[string]*Attribute{"port": {Type: cty.Number, Optional: true}},
			}},
		},
		BlockTypes: map[string]*NestedBlock{
			"timeouts": {Nesting: NestingSingle, Block: Block{Attributes: map[string]*Attribute{
				"create": {Type: cty.String, Optional: true},
			}}},
			"tag": {Nesting: NestingSet, Block: Block{Attributes: map[string]*Attribute{
				"key": {Type: cty.String, Required: true},
			}}},
			"settings": {Nesting: NestingGroup, Block: Block{Attributes: map[string]*Attribute{
				"debug": {Type: cty.Bool, Optional: true},
			}}},
			"env": {Nesting: NestingMap, Block: Block{Attributes: map[string]*Attribute{
				"value": {Type: cty.String, Optional: true},
			}}},
			"step": {Nesting: NestingList, Block: Block{Attributes: map[string]*Attribute{
				"input": {Type: cty.DynamicPseudoType, Optional: true},
			}}},
		},
	}
}

func TestImpliedType(t *testing.T) {
	got := testBlock().ImpliedType()
	want := cty.Object(map[string]cty.Type{
		"name":     cty.String,
		"count":    cty.Number,
		"arn":      cty.String,
		"dynamic":  cty.DynamicPseudoType,
		"rules":    cty.List(cty.Object(map[string]cty.Type{"port": cty.Number})),
		"timeouts": cty.Object(map[string]cty.Type{"create": cty.String}),
		"tag":      cty.Set(cty.Object(map[string]cty.Type{"key": cty.String})),
		"settings": cty.Object(map[string]cty.Type{"debug": cty.Bool}),
		"env":      cty.Map(cty.Object(map[string]cty.Type{"value": cty.String})),
		// A list of blocks with dynamically-typed content cannot be a list.
		"step": cty.DynamicPseudoType,
	})
	if !got.Equals(want) {
		t.Fatalf("implied type\n got: %#v\nwant: %#v", got, want)
	}
}

func TestCoerceValueFillsMissing(t *testing.T) {
	b := testBlock()
	got, err := b.CoerceValue(cty.ObjectVal(map[string]cty.Value{
		"name":  cty.StringVal("x"),
		"count": cty.StringVal("3"), // converted to a number
		"tag": cty.ListVal([]cty.Value{
			cty.ObjectVal(map[string]cty.Value{"key": cty.StringVal("a")}),
		}),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if errs := got.Type().TestConformance(b.ImpliedType()); len(errs) > 0 {
		t.Fatalf("coerced value does not conform to the schema: %v", errs)
	}
	checks := map[string]cty.Value{
		"name":     cty.StringVal("x"),
		"count":    cty.NumberIntVal(3),
		"arn":      cty.NullVal(cty.String),
		"timeouts": cty.NullVal(cty.Object(map[string]cty.Type{"create": cty.String})),
		"settings": cty.ObjectVal(map[string]cty.Value{"debug": cty.NullVal(cty.Bool)}),
		"env":      cty.MapValEmpty(cty.Object(map[string]cty.Type{"value": cty.String})),
		"step":     cty.EmptyTupleVal,
		"tag": cty.SetVal([]cty.Value{
			cty.ObjectVal(map[string]cty.Value{"key": cty.StringVal("a")}),
		}),
	}
	for name, want := range checks {
		if v := got.GetAttr(name); !v.RawEquals(want) {
			t.Errorf("%s = %#v, want %#v", name, v, want)
		}
	}
}

func TestCoerceValueNull(t *testing.T) {
	got, err := testBlock().CoerceValue(cty.NilVal)
	if err != nil {
		t.Fatal(err)
	}
	if got.GetAttr("name").IsKnown() && !got.GetAttr("name").IsNull() {
		t.Fatalf("name should be null, got %#v", got.GetAttr("name"))
	}
}

func TestCoerceValueRejectsUnknownArgument(t *testing.T) {
	_, err := testBlock().CoerceValue(cty.ObjectVal(map[string]cty.Value{
		"nope": cty.True,
	}))
	if err == nil {
		t.Fatal("expected an error for an unsupported argument")
	}
}
