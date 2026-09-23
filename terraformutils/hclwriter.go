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
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"github.com/Perruer/unclick/internal/schema"
	"github.com/Perruer/unclick/terraformutils/providerwrapper"
)

// ErrNoSchema means a resource was loaded from a plan file, which does not
// carry the provider schema; the legacy printer handles those.
var ErrNoSchema = errors.New("resource has no schema")

// WriteResourcesHCL renders resources as HCL using the provider schema to
// tell nested blocks from arguments. Terraformer's printer guessed from the
// data and wrote maps inside blocks as labelled blocks (for example
// `selector "match_labels" { ... }` for Kubernetes), which does not parse.
//
// String values are HCL template source, as everywhere in Item: "${...}"
// is an interpolation and "$${" a literal "${". A value that is exactly one
// interpolation, such as "${aws_vpc.main.id}", is written as a bare
// expression.
func WriteResourcesHCL(resources []Resource) ([]byte, error) {
	sorted := make([]Resource, len(resources))
	copy(sorted, resources)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].InstanceInfo.Type != sorted[j].InstanceInfo.Type {
			return sorted[i].InstanceInfo.Type < sorted[j].InstanceInfo.Type
		}
		return sorted[i].ResourceName < sorted[j].ResourceName
	})
	var b strings.Builder
	for i, r := range sorted {
		if r.Schema == nil {
			return nil, ErrNoSchema
		}
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "resource %s %s {\n", quoteLiteral(r.InstanceInfo.Type), quoteLiteral(r.ResourceName))
		writeMetaArguments(&b, r.Item)
		legacy := providerwrapper.IsLegacySDKResource(r.Schema)
		if err := writeBody(&b, r.Item, r.Schema, legacy); err != nil {
			return nil, fmt.Errorf("%s: %w", r.InstanceInfo.Id, err)
		}
		b.WriteString("}\n")
	}
	src := []byte(b.String())
	if _, diags := hclsyntax.ParseConfig(src, "generated.tf", hcl.InitialPos); diags.HasErrors() {
		return nil, fmt.Errorf("generated HCL does not parse: %s", diags.Error())
	}
	return hclwrite.Format(src), nil
}

// writeMetaArguments writes the meta-arguments some importers add to Item:
// provider (a provider configuration address) and depends_on (resource
// addresses, with or without "${...}").
func writeMetaArguments(b *strings.Builder, item map[string]interface{}) {
	if p, ok := item["provider"].(string); ok && p != "" {
		if expr := bareExpression(p); expr != "" {
			fmt.Fprintf(b, "provider = %s\n", expr)
		}
	}
	var deps []string
	switch t := item["depends_on"].(type) {
	case []string:
		deps = t
	case []interface{}:
		for _, d := range t {
			if s, ok := d.(string); ok {
				deps = append(deps, s)
			}
		}
	}
	var refs []string
	for _, d := range deps {
		if expr := bareExpression(d); expr != "" {
			refs = append(refs, expr)
		}
	}
	if len(refs) > 0 {
		fmt.Fprintf(b, "depends_on = [%s]\n", strings.Join(refs, ", "))
	}
}

// bareExpression turns "aws_x.y" or "${aws_x.y}" into a traversal that can
// be written without quotes, or returns "" if it is not one.
func bareExpression(s string) string {
	if expr, ok := singleInterpolation(s); ok {
		s = expr
	}
	if _, diags := hclsyntax.ParseTraversalAbs([]byte(s), "", hcl.InitialPos); diags.HasErrors() {
		return ""
	}
	return s
}

// writeBody writes the arguments and nested blocks of item. legacy marks
// SDKv2 resources, whose collections of objects are "attributes as blocks":
// they are written with block syntax, as in every provider example, because
// attribute syntax would need every field of every object.
func writeBody(b *strings.Builder, item map[string]interface{}, s *schema.Block, legacy bool) error {
	for _, name := range sortedKeys(s.Attributes) {
		raw, ok := item[name]
		if !ok || raw == nil {
			continue
		}
		attr := s.Attributes[name]
		ty := attr.ImpliedType()
		if legacy && attr.NestedType == nil && isObjectCollection(ty) {
			for _, elem := range blockElements(raw) {
				writeObjectBlock(b, name, elem, ty.ElementType())
			}
			continue
		}
		src, ok := valueSource(raw, ty)
		if !ok {
			continue
		}
		fmt.Fprintf(b, "%s = %s\n", name, src)
	}
	for _, name := range sortedKeys(s.BlockTypes) {
		raw, ok := item[name]
		if !ok || raw == nil {
			continue
		}
		nb := s.BlockTypes[name]
		if nb.Nesting == schema.NestingMap {
			m, _ := raw.(map[string]interface{})
			for _, key := range sortedKeys(m) {
				elem, _ := m[key].(map[string]interface{})
				fmt.Fprintf(b, "%s %s {\n", name, quoteLiteral(key))
				if err := writeBody(b, elem, &nb.Block, legacy); err != nil {
					return err
				}
				b.WriteString("}\n")
			}
			continue
		}
		for _, elem := range blockElements(raw) {
			fmt.Fprintf(b, "%s {\n", name)
			if err := writeBody(b, elem, &nb.Block, legacy); err != nil {
				return err
			}
			b.WriteString("}\n")
		}
	}
	return nil
}

func isObjectCollection(ty cty.Type) bool {
	return (ty.IsListType() || ty.IsSetType()) && ty.ElementType().IsObjectType()
}

// writeObjectBlock writes one object of an "attributes as blocks" argument.
func writeObjectBlock(b *strings.Builder, name string, elem map[string]interface{}, ty cty.Type) {
	fmt.Fprintf(b, "%s {\n", name)
	for _, k := range sortedKeys(ty.AttributeTypes()) {
		raw, ok := elem[k]
		if !ok || raw == nil {
			continue
		}
		aty := ty.AttributeType(k)
		if isObjectCollection(aty) {
			for _, inner := range blockElements(raw) {
				writeObjectBlock(b, k, inner, aty.ElementType())
			}
			continue
		}
		if src, ok := valueSource(raw, aty); ok {
			fmt.Fprintf(b, "%s = %s\n", k, src)
		}
	}
	b.WriteString("}\n")
}

// valueSource renders a value of the given type as HCL expression source.
// It reports false for values that cannot be written, which are skipped.
func valueSource(raw interface{}, ty cty.Type) (string, bool) {
	switch t := raw.(type) {
	case nil:
		return "", false
	case bool:
		return strconv.FormatBool(t), true
	case int:
		return strconv.Itoa(t), true
	case int64:
		return strconv.FormatInt(t, 10), true
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64), true
	case string:
		switch {
		case ty.Equals(cty.Number):
			if _, err := strconv.ParseFloat(t, 64); err == nil {
				return t, true
			}
		case ty.Equals(cty.Bool):
			if t == "true" || t == "false" {
				return t, true
			}
		}
		return templateSource(t), true
	case []interface{}:
		elemTy := cty.DynamicPseudoType
		if ty.IsListType() || ty.IsSetType() {
			elemTy = ty.ElementType()
		}
		parts := make([]string, 0, len(t))
		for i, e := range t {
			ety := elemTy
			if ty.IsTupleType() && i < len(ty.TupleElementTypes()) {
				ety = ty.TupleElementTypes()[i]
			}
			if s, ok := valueSource(e, ety); ok {
				parts = append(parts, s)
			}
		}
		return "[" + strings.Join(parts, ", ") + "]", true
	case map[string]interface{}:
		var b strings.Builder
		b.WriteString("{\n")
		if ty.IsObjectType() {
			// An object value needs every attribute of its type; the ones the
			// state left out are null.
			for _, k := range sortedKeys(ty.AttributeTypes()) {
				s, ok := valueSource(t[k], ty.AttributeType(k))
				if !ok {
					s = "null"
				}
				fmt.Fprintf(&b, "%s = %s\n", keySource(k), s)
			}
			b.WriteString("}")
			return b.String(), true
		}
		for _, k := range sortedKeys(t) {
			ety := cty.DynamicPseudoType
			if ty.IsMapType() {
				ety = ty.ElementType()
			}
			s, ok := valueSource(t[k], ety)
			if !ok {
				continue
			}
			fmt.Fprintf(&b, "%s = %s\n", keySource(k), s)
		}
		b.WriteString("}")
		return b.String(), true
	default:
		return templateSource(fmt.Sprint(t)), true
	}
}

// templateSource writes a string that is already HCL template source.
func templateSource(s string) string {
	if expr, ok := singleInterpolation(s); ok {
		return expr
	}
	if strings.HasSuffix(s, "\n") && !strings.Contains(s, "\nEOT\n") && !strings.HasPrefix(s, "EOT\n") {
		// A heredoc keeps multi-line documents such as policies readable. Its
		// value always ends with a newline, so it is only used when the
		// original does too.
		return "<<EOT\n" + strings.ReplaceAll(s, "%{", "%%{") + "EOT"
	}
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`, "\t", `\t`, "%{", "%%{")
	return `"` + r.Replace(s) + `"`
}

// singleInterpolation returns the expression inside a string that is
// nothing but one interpolation, such as "${aws_vpc.main.id}".
func singleInterpolation(s string) (string, bool) {
	if !strings.HasPrefix(s, "${") || !strings.HasSuffix(s, "}") {
		return "", false
	}
	inner := s[2 : len(s)-1]
	if strings.Contains(inner, "${") || strings.Contains(inner, "}") {
		return "", false
	}
	if _, diags := hclsyntax.ParseExpression([]byte(inner), "", hcl.InitialPos); diags.HasErrors() {
		return "", false
	}
	return inner, true
}

// quoteLiteral quotes a string that must be taken literally, such as a
// label or a map key.
func quoteLiteral(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`, "\t", `\t`, "${", "$${", "%{", "%%{")
	return `"` + r.Replace(s) + `"`
}

func keySource(k string) string {
	if hclsyntax.ValidIdentifier(k) {
		return k
	}
	return quoteLiteral(k)
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
