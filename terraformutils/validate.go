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
	"log"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/Perruer/unclick/internal/plugin"
	"github.com/Perruer/unclick/internal/schema"
)

// ConfigValidator is what FixInvalidConfig needs from a running provider.
type ConfigValidator interface {
	GetSchema() *schema.Provider
	ValidateResourceConfig(typeName string, config cty.Value) ([]plugin.Diagnostic, error)
}

// FixInvalidConfig asks the provider to validate every generated resource
// and drops the optional arguments it rejects. State written by the legacy
// plugin SDK holds zero values for arguments that were never set, and
// writing those back breaks rules the schema does not expose ("conflicts
// with", "all of ... must be specified"). Dropping an optional argument
// leaves it to the provider, which reads it back as the same zero value.
//
// Removed arguments are recorded in Resource.LeftOut and errors that remain
// in Resource.ProviderErrors; the HCL writer puts both in a comment above
// the resource. It returns the resources that still fail validation.
func FixInvalidConfig(resources []*Resource, v ConfigValidator) []*Resource {
	s := v.GetSchema()
	var invalid []*Resource
	for _, r := range resources {
		rs, ok := s.ResourceTypes[r.InstanceInfo.Type]
		if !ok || r.Item == nil {
			continue
		}
		fixResource(r, rs.Block, v)
		if len(r.ProviderErrors) > 0 {
			invalid = append(invalid, r)
		}
	}
	return invalid
}

// LogFixSummary reports what FixInvalidConfig could not settle, so it is
// not buried in per-resource log lines.
func LogFixSummary(invalid []*Resource) {
	if len(invalid) == 0 {
		return
	}
	ids := make([]string, 0, len(invalid))
	for _, r := range invalid {
		ids = append(ids, r.InstanceInfo.Type+"."+r.ResourceName)
	}
	log.Printf("WARN: %d resources still fail provider validation; the errors are in a comment above each of them: %s",
		len(invalid), strings.Join(ids, ", "))
}

func fixResource(r *Resource, b *schema.Block, v ConfigValidator) {
	// Every round removes one argument that is present, so the loop ends
	// after at most that many removals and one last validation.
	limit := countOptional(r.Item, b)
	for round := 0; ; round++ {
		config, err := ItemValue(r.Item, b)
		if err != nil {
			log.Printf("cannot validate %s: %v", r.InstanceInfo.Id, err)
			return
		}
		diags, err := v.ValidateResourceConfig(r.InstanceInfo.Type, config)
		if err != nil {
			log.Printf("cannot validate %s: %v", r.InstanceInfo.Id, err)
			return
		}
		var errs []plugin.Diagnostic
		for _, d := range diags {
			if d.Error {
				errs = append(errs, d)
			}
		}
		if len(errs) == 0 {
			return
		}

		// Remove one argument per round: two arguments that conflict are
		// both reported, and dropping either one settles it. Prefer one that
		// holds a zero value, which the legacy SDK writes for arguments that
		// were never set; an explicit zero could not have been in the user's
		// configuration next to the argument it conflicts with.
		var pick *plugin.Diagnostic
		var pickTargets []optionalTarget
		if round < limit {
			for i := range errs {
				targets := findOptional(r.Item, b, errs[i].Path)
				if len(targets) == 0 {
					continue
				}
				if pick == nil || (allZero(targets) && !allZero(pickTargets)) {
					pick, pickTargets = &errs[i], targets
				}
			}
		}
		if pick == nil {
			for _, d := range errs {
				msg := d.Summary
				if d.Detail != "" {
					msg += ": " + d.Detail
				}
				if p := pathString(d.Path); p != "" {
					msg = p + ": " + msg
				}
				r.ProviderErrors = append(r.ProviderErrors, msg)
				log.Printf("WARN: %s: %s", r.InstanceInfo.Id, msg)
			}
			return
		}
		for _, t := range pickTargets {
			delete(t.item, t.name)
		}
		left := fmt.Sprintf("%s (%s)", pathString(pick.Path), pick.Summary)
		r.LeftOut = append(r.LeftOut, left)
		log.Printf("%s: left out %s", r.InstanceInfo.Id, left)
	}
}

// optionalTarget is one optional argument in a resource body.
type optionalTarget struct {
	item map[string]interface{}
	name string
}

// findOptional returns the arguments a diagnostic points at, if the schema
// says they are optional and they are present.
func findOptional(item map[string]interface{}, b *schema.Block, path []plugin.PathStep) []optionalTarget {
	if len(path) == 0 || path[0].Attribute == "" {
		return nil
	}
	name := path[0].Attribute
	if attr, ok := b.Attributes[name]; ok {
		if len(path) > 1 || attr.Required || !attr.Optional {
			return nil
		}
		if _, present := item[name]; !present {
			return nil
		}
		return []optionalTarget{{item, name}}
	}
	nb, ok := b.BlockTypes[name]
	if !ok {
		return nil
	}
	rest := path[1:]
	var index int64 = -1
	if len(rest) > 0 && rest[0].IsIndex {
		index = rest[0].Index
		rest = rest[1:]
	}
	var out []optionalTarget
	for i, elem := range blockElements(item[name]) {
		if index >= 0 && int64(i) != index {
			continue
		}
		out = append(out, findOptional(elem, &nb.Block, rest)...)
	}
	return out
}

// countOptional counts the optional arguments present in a body, including
// those of nested blocks: the most FixInvalidConfig can ever remove.
func countOptional(item map[string]interface{}, b *schema.Block) int {
	n := 0
	for name, attr := range b.Attributes {
		if _, present := item[name]; present && attr.Optional && !attr.Required {
			n++
		}
	}
	for name, nb := range b.BlockTypes {
		for _, elem := range blockElements(item[name]) {
			n += countOptional(elem, &nb.Block)
		}
	}
	return n
}

func allZero(targets []optionalTarget) bool {
	for _, t := range targets {
		if !isZeroValue(t.item[t.name]) {
			return false
		}
	}
	return len(targets) > 0
}

// isZeroValue reports whether a body value is what the legacy SDK writes for
// an argument that was never set.
func isZeroValue(v interface{}) bool {
	switch t := v.(type) {
	case nil:
		return true
	case string:
		return t == "" || t == "0" || t == "false"
	case bool:
		return !t
	case int:
		return t == 0
	case int64:
		return t == 0
	case float64:
		return t == 0
	case []interface{}:
		return len(t) == 0
	case map[string]interface{}:
		return len(t) == 0
	default:
		return false
	}
}

func blockElements(v interface{}) []map[string]interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		return []map[string]interface{}{t}
	case []interface{}:
		var out []map[string]interface{}
		for _, e := range t {
			if m, ok := e.(map[string]interface{}); ok {
				out = append(out, m)
			}
		}
		return out
	case []map[string]interface{}:
		return t
	default:
		return nil
	}
}

func pathString(path []plugin.PathStep) string {
	var b strings.Builder
	for _, p := range path {
		switch {
		case p.IsIndex:
			fmt.Fprintf(&b, "[%d]", p.Index)
		case p.IsKey:
			fmt.Fprintf(&b, "[%q]", p.Key)
		default:
			if b.Len() > 0 {
				b.WriteByte('.')
			}
			b.WriteString(p.Attribute)
		}
	}
	return b.String()
}

// ItemValue converts a generated resource body into a value of the schema's
// type, the form providers validate. Interpolations such as
// "${aws_vpc.main.id}" become unknown values, as they are during `plan`.
func ItemValue(item map[string]interface{}, b *schema.Block) (cty.Value, error) {
	vals := make(map[string]cty.Value, len(b.Attributes)+len(b.BlockTypes))
	for name, attr := range b.Attributes {
		ty := attr.ImpliedType()
		raw, ok := item[name]
		if !ok || raw == nil {
			vals[name] = cty.NullVal(ty)
			continue
		}
		v, err := itemToValue(raw, ty)
		if err != nil {
			return cty.NilVal, fmt.Errorf("%s: %w", name, err)
		}
		vals[name] = v
	}
	for name, nb := range b.BlockTypes {
		v, err := nestedValue(item[name], nb)
		if err != nil {
			return cty.NilVal, fmt.Errorf("%s: %w", name, err)
		}
		vals[name] = v
	}
	return cty.ObjectVal(vals), nil
}

func nestedValue(raw interface{}, nb *schema.NestedBlock) (cty.Value, error) {
	elemTy := nb.Block.ImpliedType()
	var elems []cty.Value
	for _, e := range blockElements(raw) {
		v, err := ItemValue(e, &nb.Block)
		if err != nil {
			return cty.NilVal, err
		}
		elems = append(elems, v)
	}
	switch nb.Nesting {
	case schema.NestingSingle:
		if len(elems) == 0 {
			return cty.NullVal(elemTy), nil
		}
		return elems[0], nil
	case schema.NestingGroup:
		if len(elems) == 0 {
			return ItemValue(nil, &nb.Block)
		}
		return elems[0], nil
	case schema.NestingSet:
		if len(elems) == 0 {
			return cty.SetValEmpty(elemTy), nil
		}
		return cty.SetVal(elems), nil
	case schema.NestingMap:
		// Terraformer never produces map-nested blocks; treat them as absent.
		return cty.MapValEmpty(elemTy), nil
	default:
		if elemTy.HasDynamicTypes() {
			return cty.TupleVal(elems), nil
		}
		if len(elems) == 0 {
			return cty.ListValEmpty(elemTy), nil
		}
		return cty.ListVal(elems), nil
	}
}

func itemToValue(raw interface{}, ty cty.Type) (cty.Value, error) {
	if s, ok := raw.(string); ok && strings.Contains(s, "${") {
		return cty.UnknownVal(ty), nil
	}
	switch {
	case ty == cty.DynamicPseudoType:
		if s, ok := raw.(string); ok {
			return cty.StringVal(s), nil
		}
		return cty.DynamicVal, nil
	case ty.IsPrimitiveType():
		var v cty.Value
		switch t := raw.(type) {
		case string:
			v = cty.StringVal(t)
		case bool, int, int64, float64:
			impl, err := gocty.ImpliedType(t)
			if err != nil {
				return cty.NilVal, err
			}
			if v, err = gocty.ToCtyValue(t, impl); err != nil {
				return cty.NilVal, err
			}
		default:
			return cty.NilVal, fmt.Errorf("unexpected %T for %s", raw, ty.FriendlyName())
		}
		return convert.Convert(v, ty)
	case ty.IsListType() || ty.IsSetType():
		list, ok := raw.([]interface{})
		if !ok {
			return cty.NilVal, fmt.Errorf("expected a list, got %T", raw)
		}
		elems := make([]cty.Value, 0, len(list))
		for _, e := range list {
			v, err := itemToValue(e, ty.ElementType())
			if err != nil {
				return cty.NilVal, err
			}
			elems = append(elems, v)
		}
		if ty.IsSetType() {
			if len(elems) == 0 {
				return cty.SetValEmpty(ty.ElementType()), nil
			}
			return cty.SetVal(elems), nil
		}
		if len(elems) == 0 {
			return cty.ListValEmpty(ty.ElementType()), nil
		}
		return cty.ListVal(elems), nil
	case ty.IsMapType():
		m, ok := raw.(map[string]interface{})
		if !ok {
			return cty.NilVal, fmt.Errorf("expected a map, got %T", raw)
		}
		if len(m) == 0 {
			return cty.MapValEmpty(ty.ElementType()), nil
		}
		elems := make(map[string]cty.Value, len(m))
		for k, e := range m {
			v, err := itemToValue(e, ty.ElementType())
			if err != nil {
				return cty.NilVal, err
			}
			elems[k] = v
		}
		return cty.MapVal(elems), nil
	case ty.IsObjectType():
		m, ok := raw.(map[string]interface{})
		if !ok {
			if l, isList := raw.([]interface{}); isList && len(l) == 1 {
				m, ok = l[0].(map[string]interface{})
			}
		}
		if !ok {
			return cty.NilVal, fmt.Errorf("expected an object, got %T", raw)
		}
		attrs := make(map[string]cty.Value, len(ty.AttributeTypes()))
		for name, aty := range ty.AttributeTypes() {
			e, present := m[name]
			if !present || e == nil {
				attrs[name] = cty.NullVal(aty)
				continue
			}
			v, err := itemToValue(e, aty)
			if err != nil {
				return cty.NilVal, err
			}
			attrs[name] = v
		}
		return cty.ObjectVal(attrs), nil
	default:
		return cty.NilVal, fmt.Errorf("unsupported type %s", ty.FriendlyName())
	}
}
