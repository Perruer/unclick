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
	"context"
	"fmt"

	"github.com/zclconf/go-cty/cty"

	"github.com/Perruer/unclick/internal/schema"
	proto "github.com/Perruer/unclick/internal/tfplugin5"
)

type v5 struct {
	client proto.ProviderClient
}

var caps5 = &proto.ClientCapabilities{}

func (p *v5) getSchema(ctx context.Context) (*schema.Provider, error) {
	resp, err := p.client.GetSchema(ctx, &proto.GetProviderSchema_Request{})
	if err != nil {
		return nil, fmt.Errorf("GetProviderSchema: %w", err)
	}
	if err := diagError("GetProviderSchema", diags5(resp.Diagnostics)); err != nil {
		return nil, err
	}
	out := &schema.Provider{
		ResourceTypes: map[string]schema.Resource{},
		DataSources:   map[string]schema.Resource{},
		ListResources: map[string]schema.Resource{},
	}
	if out.Provider, err = block5(resp.Provider.GetBlock()); err != nil {
		return nil, fmt.Errorf("provider schema: %w", err)
	}
	for _, set := range []struct {
		dst map[string]schema.Resource
		src map[string]*proto.Schema
	}{
		{out.ResourceTypes, resp.ResourceSchemas},
		{out.DataSources, resp.DataSourceSchemas},
		{out.ListResources, resp.ListResourceSchemas},
	} {
		for name, s := range set.src {
			b, err := block5(s.GetBlock())
			if err != nil {
				return nil, fmt.Errorf("schema of %s: %w", name, err)
			}
			set.dst[name] = schema.Resource{Version: s.GetVersion(), Block: b}
		}
	}
	return out, nil
}

func (p *v5) configure(ctx context.Context, config cty.Value, ty cty.Type) error {
	raw, err := encode(config, ty)
	if err != nil {
		return fmt.Errorf("encoding provider configuration: %w", err)
	}
	// Protocol 5 providers may fill in defaults while validating, and expect
	// to be configured with the result.
	prep, err := p.client.PrepareProviderConfig(ctx, &proto.PrepareProviderConfig_Request{
		Config: &proto.DynamicValue{Msgpack: raw},
	})
	if err != nil {
		return fmt.Errorf("PrepareProviderConfig: %w", err)
	}
	if err := diagError("PrepareProviderConfig", diags5(prep.Diagnostics)); err != nil {
		return err
	}
	if pc := prep.GetPreparedConfig(); pc != nil && (len(pc.Msgpack) > 0 || len(pc.Json) > 0) {
		prepared, err := decode(pc.Msgpack, pc.Json, ty)
		if err != nil {
			return fmt.Errorf("decoding prepared provider configuration: %w", err)
		}
		if raw, err = encode(prepared, ty); err != nil {
			return fmt.Errorf("encoding prepared provider configuration: %w", err)
		}
	}
	resp, err := p.client.Configure(ctx, &proto.Configure_Request{
		TerraformVersion:   TerraformVersion,
		Config:             &proto.DynamicValue{Msgpack: raw},
		ClientCapabilities: caps5,
	})
	if err != nil {
		return fmt.Errorf("ConfigureProvider: %w", err)
	}
	return diagError("ConfigureProvider", diags5(resp.Diagnostics))
}

func (p *v5) readResource(ctx context.Context, typeName string, state cty.Value, ty cty.Type, private []byte) (cty.Value, []byte, error) {
	raw, err := encode(state, ty)
	if err != nil {
		return cty.NilVal, nil, fmt.Errorf("encoding %s state: %w", typeName, err)
	}
	resp, err := p.client.ReadResource(ctx, &proto.ReadResource_Request{
		TypeName:           typeName,
		CurrentState:       &proto.DynamicValue{Msgpack: raw},
		Private:            private,
		ClientCapabilities: caps5,
	})
	if err != nil {
		return cty.NilVal, nil, fmt.Errorf("ReadResource %s: %w", typeName, err)
	}
	if err := diagError("ReadResource "+typeName, diags5(resp.Diagnostics)); err != nil {
		return cty.NilVal, nil, err
	}
	v, err := decode(resp.GetNewState().GetMsgpack(), resp.GetNewState().GetJson(), ty)
	if err != nil {
		return cty.NilVal, nil, fmt.Errorf("decoding %s state: %w", typeName, err)
	}
	return v, resp.Private, nil
}

func (p *v5) importResourceState(ctx context.Context, typeName, id string, types func(string) (cty.Type, bool)) ([]ImportedResource, error) {
	resp, err := p.client.ImportResourceState(ctx, &proto.ImportResourceState_Request{
		TypeName:           typeName,
		Id:                 id,
		ClientCapabilities: caps5,
	})
	if err != nil {
		return nil, fmt.Errorf("ImportResourceState %s: %w", typeName, err)
	}
	if err := diagError("ImportResourceState "+typeName, diags5(resp.Diagnostics)); err != nil {
		return nil, err
	}
	var out []ImportedResource
	for _, r := range resp.ImportedResources {
		ty, ok := types(r.TypeName)
		if !ok {
			return nil, fmt.Errorf("import of %s returned unknown type %q", typeName, r.TypeName)
		}
		v, err := decode(r.GetState().GetMsgpack(), r.GetState().GetJson(), ty)
		if err != nil {
			return nil, fmt.Errorf("decoding imported %s: %w", r.TypeName, err)
		}
		out = append(out, ImportedResource{TypeName: r.TypeName, State: v, Private: r.Private})
	}
	return out, nil
}

func (p *v5) validateResourceConfig(ctx context.Context, typeName string, config cty.Value, ty cty.Type) ([]Diagnostic, error) {
	raw, err := encode(config, ty)
	if err != nil {
		return nil, fmt.Errorf("encoding %s configuration: %w", typeName, err)
	}
	resp, err := p.client.ValidateResourceTypeConfig(ctx, &proto.ValidateResourceTypeConfig_Request{
		TypeName:           typeName,
		Config:             &proto.DynamicValue{Msgpack: raw},
		ClientCapabilities: caps5,
	})
	if err != nil {
		return nil, fmt.Errorf("ValidateResourceTypeConfig %s: %w", typeName, err)
	}
	return diags5(resp.Diagnostics), nil
}

func diags5(ds []*proto.Diagnostic) []Diagnostic {
	out := make([]Diagnostic, 0, len(ds))
	for _, d := range ds {
		diag := Diagnostic{
			Error:   d.Severity == proto.Diagnostic_ERROR,
			Summary: d.Summary,
			Detail:  d.Detail,
		}
		for _, step := range d.GetAttribute().GetSteps() {
			switch sel := step.Selector.(type) {
			case *proto.AttributePath_Step_AttributeName:
				diag.Path = append(diag.Path, PathStep{Attribute: sel.AttributeName})
			case *proto.AttributePath_Step_ElementKeyString:
				diag.Path = append(diag.Path, PathStep{Key: sel.ElementKeyString, IsKey: true})
			case *proto.AttributePath_Step_ElementKeyInt:
				diag.Path = append(diag.Path, PathStep{Index: sel.ElementKeyInt, IsIndex: true})
			}
		}
		out = append(out, diag)
	}
	return out
}

func block5(b *proto.Schema_Block) (*schema.Block, error) {
	out := &schema.Block{
		Attributes: map[string]*schema.Attribute{},
		BlockTypes: map[string]*schema.NestedBlock{},
	}
	if b == nil {
		return out, nil
	}
	out.Deprecated = b.Deprecated
	for _, a := range b.Attributes {
		ty, err := decodeType(a.Type)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", a.Name, err)
		}
		out.Attributes[a.Name] = &schema.Attribute{
			Type:       ty,
			Required:   a.Required,
			Optional:   a.Optional,
			Computed:   a.Computed,
			Sensitive:  a.Sensitive,
			Deprecated: a.Deprecated,
			WriteOnly:  a.WriteOnly,
		}
	}
	for _, nb := range b.BlockTypes {
		inner, err := block5(nb.Block)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", nb.TypeName, err)
		}
		nesting, err := nesting5(nb.Nesting)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", nb.TypeName, err)
		}
		out.BlockTypes[nb.TypeName] = &schema.NestedBlock{
			Block:    *inner,
			Nesting:  nesting,
			MinItems: int(nb.MinItems),
			MaxItems: int(nb.MaxItems),
		}
	}
	return out, nil
}

func nesting5(n proto.Schema_NestedBlock_NestingMode) (schema.NestingMode, error) {
	switch n {
	case proto.Schema_NestedBlock_SINGLE:
		return schema.NestingSingle, nil
	case proto.Schema_NestedBlock_GROUP:
		return schema.NestingGroup, nil
	case proto.Schema_NestedBlock_LIST:
		return schema.NestingList, nil
	case proto.Schema_NestedBlock_SET:
		return schema.NestingSet, nil
	case proto.Schema_NestedBlock_MAP:
		return schema.NestingMap, nil
	default:
		return 0, fmt.Errorf("unsupported block nesting %v", n)
	}
}
