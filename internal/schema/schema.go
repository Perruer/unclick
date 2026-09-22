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

// Package schema describes provider schemas independently of the plugin
// protocol version (5 or 6) the provider speaks.
package schema

import (
	"github.com/zclconf/go-cty/cty"
)

// NestingMode says how a nested block or a nested attribute object repeats.
type NestingMode int

const (
	NestingSingle NestingMode = iota
	NestingGroup
	NestingList
	NestingSet
	NestingMap
)

// Block is the body of a resource, a data source, a provider configuration
// or a nested block.
type Block struct {
	Attributes map[string]*Attribute
	BlockTypes map[string]*NestedBlock
	Deprecated bool
}

// Attribute is a single argument of a block. Exactly one of Type and
// NestedType is set; NestedType is only used by protocol 6 providers.
type Attribute struct {
	Type       cty.Type
	NestedType *Object

	Required   bool
	Optional   bool
	Computed   bool
	Sensitive  bool
	Deprecated bool
	WriteOnly  bool
}

// Object is the type of a protocol 6 nested attribute.
type Object struct {
	Attributes map[string]*Attribute
	Nesting    NestingMode
}

// NestedBlock is a block type that can appear inside another block.
type NestedBlock struct {
	Block
	Nesting  NestingMode
	MinItems int
	MaxItems int
}

// Resource is the schema of a managed resource, data source or list
// resource type.
type Resource struct {
	Version int64
	Block   *Block
}

// Provider is everything a provider reports about itself.
type Provider struct {
	Provider      *Block
	ResourceTypes map[string]Resource
	DataSources   map[string]Resource
	ListResources map[string]Resource
}

// IsConfigurable reports whether the attribute can be written in
// configuration, as opposed to being set only by the provider.
func (a *Attribute) IsConfigurable() bool {
	return a.Required || a.Optional
}

// ImpliedType is the cty type of values that conform to the block.
func (b *Block) ImpliedType() cty.Type {
	if b == nil {
		return cty.EmptyObject
	}
	attrs := make(map[string]cty.Type, len(b.Attributes)+len(b.BlockTypes))
	for name, attr := range b.Attributes {
		attrs[name] = attr.ImpliedType()
	}
	for name, nb := range b.BlockTypes {
		attrs[name] = nb.ImpliedType()
	}
	return cty.Object(attrs)
}

// ImpliedType is the cty type of values that conform to the attribute.
func (a *Attribute) ImpliedType() cty.Type {
	if a.NestedType != nil {
		return a.NestedType.ImpliedType()
	}
	return a.Type
}

// ImpliedType is the cty type of values that conform to the nested object.
func (o *Object) ImpliedType() cty.Type {
	attrs := make(map[string]cty.Type, len(o.Attributes))
	for name, attr := range o.Attributes {
		attrs[name] = attr.ImpliedType()
	}
	return wrap(cty.Object(attrs), o.Nesting)
}

// ImpliedType is the cty type of the collection of nested blocks.
func (nb *NestedBlock) ImpliedType() cty.Type {
	return wrap(nb.Block.ImpliedType(), nb.Nesting)
}

func wrap(elem cty.Type, nesting NestingMode) cty.Type {
	switch nesting {
	case NestingList:
		// A list whose elements may differ in type cannot be a cty list,
		// so Terraform represents it as a dynamically-typed value.
		if elem.HasDynamicTypes() {
			return cty.DynamicPseudoType
		}
		return cty.List(elem)
	case NestingSet:
		return cty.Set(elem)
	case NestingMap:
		if elem.HasDynamicTypes() {
			return cty.DynamicPseudoType
		}
		return cty.Map(elem)
	default:
		return elem
	}
}
