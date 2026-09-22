package providerwrapper //nolint

import (
	"regexp"
	"testing"

	"github.com/Perruer/unclick/internal/schema"
	"github.com/zclconf/go-cty/cty"
)

func TestIgnoredAttributes(t *testing.T) {
	attributes := map[string]*schema.Attribute{
		"computed_attribute": {
			Type:     cty.Number,
			Computed: true,
		},
		"required_attribute": {
			Type:     cty.String,
			Required: true,
		},
	}

	testCases := map[string]struct {
		block                map[string]*schema.NestedBlock
		ignoredAttributes    []string
		notIgnoredAttributes []string
	}{
		"nesting_set": {map[string]*schema.NestedBlock{
			"attribute_one": {
				Block: schema.Block{
					Attributes: attributes,
				},
				Nesting: schema.NestingSet,
			},
		}, []string{"nesting_set.attribute_one.computed_attribute"},
			[]string{"nesting_set.attribute_one.required_attribute"}},
		"nesting_list": {map[string]*schema.NestedBlock{
			"attribute_one": {
				Block: schema.Block{
					Attributes: map[string]*schema.Attribute{},
					BlockTypes: map[string]*schema.NestedBlock{
						"attribute_two_nested": {
							Nesting: schema.NestingList,
							Block: schema.Block{
								Attributes: attributes,
							},
						},
					},
				},
				Nesting: schema.NestingList,
			},
		}, []string{"nesting_list.0.attribute_one.0.attribute_two_nested.computed_attribute"},
			[]string{"nesting_list.0.attribute_one.0.attribute_two_nested.required_attribute"}},
	}

	for key, tc := range testCases {
		t.Run(key, func(t *testing.T) {
			readOnlyAttributes := readObjBlocks(tc.block, []string{}, key)
			for _, attr := range tc.ignoredAttributes {
				if ignored := isAttributeIgnored(attr, readOnlyAttributes); !ignored {
					t.Errorf("attribute \"%s\" was not ignored. Pattern list: %s", attr, readOnlyAttributes)
				}
			}

			for _, attr := range tc.notIgnoredAttributes {
				if ignored := isAttributeIgnored(attr, readOnlyAttributes); ignored {
					t.Errorf("attribute \"%s\" was ignored. Pattern list: %s", attr, readOnlyAttributes)
				}
			}
		})
	}
}

func isAttributeIgnored(name string, patterns []string) bool {
	ignored := false
	for _, pattern := range patterns {
		if match, _ := regexp.MatchString(pattern, name); match {
			ignored = true
			break
		}
	}
	return ignored
}

func TestDeprecatedAttributesAreIgnored(t *testing.T) {
	s := &schema.Provider{ResourceTypes: map[string]schema.Resource{
		"github_repository": {Block: &schema.Block{
			Attributes: map[string]*schema.Attribute{
				"name":       {Type: cty.String, Required: true},
				"visibility": {Type: cty.String, Optional: true, Computed: true},
				"private":    {Type: cty.Bool, Optional: true, Computed: true, Deprecated: true},
				"etag":       {Type: cty.String, Computed: true},
			},
			BlockTypes: map[string]*schema.NestedBlock{
				"template": {Nesting: schema.NestingList, Block: schema.Block{
					Attributes: map[string]*schema.Attribute{"owner": {Type: cty.String, Required: true}},
				}},
				"old_pages": {Nesting: schema.NestingList, Block: schema.Block{
					Deprecated: true,
					Attributes: map[string]*schema.Attribute{"branch": {Type: cty.String, Optional: true}},
				}},
			},
		}},
	}}
	patterns := readOnlyAttributesOf(s, []string{"github_repository"})["github_repository"]
	for _, key := range []string{"id", "private", "etag", "old_pages.#", "old_pages.0.branch"} {
		if !isAttributeIgnored(key, patterns) {
			t.Errorf("%s should be ignored; patterns: %v", key, patterns)
		}
	}
	for _, key := range []string{"name", "visibility", "template.0.owner"} {
		if isAttributeIgnored(key, patterns) {
			t.Errorf("%s should be kept; patterns: %v", key, patterns)
		}
	}
}

func TestZeroNumbersOnlyForLegacySDKResources(t *testing.T) {
	sdkv2 := &schema.Block{
		Attributes: map[string]*schema.Attribute{
			"id":                  {Type: cty.String, Optional: true, Computed: true},
			"ipv6_netmask_length": {Type: cty.Number, Optional: true},
			"from_port":           {Type: cty.Number, Required: true},
			"enable_dns":          {Type: cty.Bool, Optional: true},
		},
		BlockTypes: map[string]*schema.NestedBlock{
			"rule": {Nesting: schema.NestingSet, Block: schema.Block{
				Attributes: map[string]*schema.Attribute{"priority": {Type: cty.Number, Optional: true}},
			}},
		},
	}
	framework := &schema.Block{Attributes: map[string]*schema.Attribute{
		"id":    {Type: cty.String, Computed: true},
		"count": {Type: cty.Number, Optional: true},
	}}
	s := &schema.Provider{ResourceTypes: map[string]schema.Resource{
		"aws_vpc":   {Block: sdkv2},
		"new_thing": {Block: framework},
	}}
	got := zeroNumberAttributesOf(s, []string{"aws_vpc", "new_thing"})
	if _, ok := got["new_thing"]; ok {
		t.Error("framework resources store null for unset numbers and must be left alone")
	}
	patterns := got["aws_vpc"]
	for _, key := range []string{"ipv6_netmask_length", "rule.2837491.priority"} {
		if !isAttributeIgnored(key, patterns) {
			t.Errorf("%s should be a zero-number key; patterns: %v", key, patterns)
		}
	}
	for _, key := range []string{"from_port", "enable_dns", "id"} {
		if isAttributeIgnored(key, patterns) {
			t.Errorf("%s must not be a zero-number key; patterns: %v", key, patterns)
		}
	}
}
