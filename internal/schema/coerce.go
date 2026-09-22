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
	"fmt"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

// CoerceValue turns a partial object, such as a provider configuration built
// from command line flags, into a value of the block's implied type: missing
// attributes become null and missing nested blocks become empty.
func (b *Block) CoerceValue(in cty.Value) (cty.Value, error) {
	return b.coerce(in, cty.Path{})
}

func (b *Block) coerce(in cty.Value, path cty.Path) (cty.Value, error) {
	if b == nil {
		return cty.EmptyObjectVal, nil
	}
	if in == cty.NilVal || in.IsNull() {
		in = cty.EmptyObjectVal
	}
	if !in.IsKnown() {
		return cty.UnknownVal(b.ImpliedType()), nil
	}
	if !in.Type().IsObjectType() && !in.Type().IsMapType() {
		return cty.NilVal, path.NewErrorf("an object is required")
	}

	given := map[string]cty.Value{}
	for it := in.ElementIterator(); it.Next(); {
		k, v := it.Element()
		given[k.AsString()] = v
	}
	for name := range given {
		if b.Attributes[name] == nil && b.BlockTypes[name] == nil {
			return cty.NilVal, path.NewErrorf("unsupported argument %q", name)
		}
	}

	out := make(map[string]cty.Value, len(b.Attributes)+len(b.BlockTypes))
	for name, attr := range b.Attributes {
		ty := attr.ImpliedType()
		v, ok := given[name]
		if !ok || v.IsNull() {
			out[name] = cty.NullVal(ty)
			continue
		}
		cv, err := convert.Convert(v, ty)
		if err != nil {
			return cty.NilVal, path.GetAttr(name).NewError(err)
		}
		out[name] = cv
	}
	for name, nb := range b.BlockTypes {
		v, err := nb.coerce(given[name], path.GetAttr(name))
		if err != nil {
			return cty.NilVal, err
		}
		out[name] = v
	}
	return cty.ObjectVal(out), nil
}

func (nb *NestedBlock) coerce(in cty.Value, path cty.Path) (cty.Value, error) {
	elemTy := nb.Block.ImpliedType()
	absent := in == cty.NilVal || in.IsNull()

	switch nb.Nesting {
	case NestingSingle:
		if absent {
			return cty.NullVal(elemTy), nil
		}
		return nb.Block.coerce(in, path)
	case NestingGroup:
		// A group always exists; its attributes are simply null when unset.
		return nb.Block.coerce(in, path)
	case NestingList, NestingSet:
		if absent {
			return emptyCollection(nb.Nesting, elemTy), nil
		}
		if !in.CanIterateElements() {
			return cty.NilVal, path.NewErrorf("a list of blocks is required")
		}
		var elems []cty.Value
		for it := in.ElementIterator(); it.Next(); {
			_, ev := it.Element()
			cv, err := nb.Block.coerce(ev, path)
			if err != nil {
				return cty.NilVal, err
			}
			elems = append(elems, cv)
		}
		if len(elems) == 0 {
			return emptyCollection(nb.Nesting, elemTy), nil
		}
		if nb.Nesting == NestingSet {
			return cty.SetVal(elems), nil
		}
		if elemTy.HasDynamicTypes() {
			return cty.TupleVal(elems), nil
		}
		return cty.ListVal(elems), nil
	case NestingMap:
		if absent {
			return cty.MapValEmpty(elemTy), nil
		}
		elems := map[string]cty.Value{}
		for it := in.ElementIterator(); it.Next(); {
			k, ev := it.Element()
			cv, err := nb.Block.coerce(ev, path.Index(k))
			if err != nil {
				return cty.NilVal, err
			}
			elems[k.AsString()] = cv
		}
		if len(elems) == 0 {
			return cty.MapValEmpty(elemTy), nil
		}
		if elemTy.HasDynamicTypes() {
			return cty.ObjectVal(elems), nil
		}
		return cty.MapVal(elems), nil
	default:
		return cty.NilVal, fmt.Errorf("unknown nesting mode %d", nb.Nesting)
	}
}

func emptyCollection(nesting NestingMode, elemTy cty.Type) cty.Value {
	if nesting == NestingSet {
		return cty.SetValEmpty(elemTy)
	}
	if elemTy.HasDynamicTypes() {
		return cty.EmptyTupleVal
	}
	return cty.ListValEmpty(elemTy)
}
