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
