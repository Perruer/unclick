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

package plugin

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
	"github.com/zclconf/go-cty/cty/msgpack"
)

// encode serializes a value the way providers expect it in a DynamicValue.
func encode(v cty.Value, ty cty.Type) ([]byte, error) {
	if v == cty.NilVal {
		v = cty.NullVal(ty)
	}
	return msgpack.Marshal(v, ty)
}

// decode reads a DynamicValue; providers may answer in msgpack or JSON.
func decode(mp, js []byte, ty cty.Type) (cty.Value, error) {
	switch {
	case len(mp) > 0:
		return msgpack.Unmarshal(mp, ty)
	case len(js) > 0:
		return ctyjson.Unmarshal(js, ty)
	default:
		return cty.NullVal(ty), nil
	}
}

// decodeType parses the JSON type description used in schema attributes.
func decodeType(raw []byte) (cty.Type, error) {
	if len(raw) == 0 {
		return cty.DynamicPseudoType, nil
	}
	var ty cty.Type
	if err := ty.UnmarshalJSON(raw); err != nil {
		return cty.NilType, fmt.Errorf("attribute type %s: %w", raw, err)
	}
	return ty, nil
}

// diagnostic is the protocol-independent part of a provider diagnostic.
type diagnostic struct {
	isError bool
	summary string
	detail  string
}

// diagError joins error diagnostics into one error; warnings are dropped
// because they rarely matter when reading existing objects.
func diagError(what string, diags []diagnostic) error {
	var msgs []string
	for _, d := range diags {
		if !d.isError {
			continue
		}
		msg := d.summary
		if d.detail != "" {
			msg += ": " + d.detail
		}
		msgs = append(msgs, msg)
	}
	if len(msgs) == 0 {
		return nil
	}
	return errors.New(what + ": " + strings.Join(msgs, "; "))
}
