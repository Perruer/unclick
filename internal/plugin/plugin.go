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

// Package plugin starts provider plugins built for OpenTofu or Terraform and
// talks to them over plugin protocol 5 or 6, whichever the provider speaks.
package plugin

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sync"

	"github.com/hashicorp/go-hclog"
	goplugin "github.com/hashicorp/go-plugin"
	"github.com/zclconf/go-cty/cty"
	"google.golang.org/grpc"

	"github.com/Perruer/unclick/internal/schema"
	"github.com/Perruer/unclick/internal/tfplugin5"
	"github.com/Perruer/unclick/internal/tfplugin6"
)

// TerraformVersion is reported to providers when they are configured. Some
// providers put it into their User-Agent or refuse very old versions.
var TerraformVersion = "1.12.0"

// maxRecvMsgSize allows for provider schemas far larger than gRPC's 4 MB
// default; the AWS provider alone exceeds it.
const maxRecvMsgSize = 256 << 20

// handshake must match what OpenTofu and Terraform send, otherwise the
// provider refuses to start and prints a "this binary is a plugin" message.
var handshake = goplugin.HandshakeConfig{
	ProtocolVersion:  4,
	MagicCookieKey:   "TF_PLUGIN_MAGIC_COOKIE",
	MagicCookieValue: "d602bf8f470bc67ca7faa0386276bbdd4330efaf76d1a219cb4d6991ca9872b2",
}

var versionedPlugins = map[int]goplugin.PluginSet{
	5: {"provider": &grpcPlugin{version: 5}},
	6: {"provider": &grpcPlugin{version: 6}},
}

type grpcPlugin struct {
	goplugin.NetRPCUnsupportedPlugin
	version int
}

func (p *grpcPlugin) GRPCServer(*goplugin.GRPCBroker, *grpc.Server) error {
	return errors.New("unclick only acts as a plugin client")
}

func (p *grpcPlugin) GRPCClient(_ context.Context, _ *goplugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	if p.version == 6 {
		return &v6{client: tfplugin6.NewProviderClient(conn)}, nil
	}
	return &v5{client: tfplugin5.NewProviderClient(conn)}, nil
}

// ImportedResource is one object returned by ImportResourceState. Importing
// a single ID can yield several objects, for example a security group and its
// rules.
type ImportedResource struct {
	TypeName string
	State    cty.Value
	Private  []byte
}

// protocol is implemented once per plugin protocol version.
type protocol interface {
	getSchema(ctx context.Context) (*schema.Provider, error)
	configure(ctx context.Context, config cty.Value, ty cty.Type) error
	readResource(ctx context.Context, typeName string, state cty.Value, ty cty.Type, private []byte) (cty.Value, []byte, error)
	importResourceState(ctx context.Context, typeName, id string, types func(string) (cty.Type, bool)) ([]ImportedResource, error)
	validateResourceConfig(ctx context.Context, typeName string, config cty.Value, ty cty.Type) ([]Diagnostic, error)
	listResource(ctx context.Context, typeName string, config cty.Value, configTy cty.Type, includeResource bool, limit int64, objectTy cty.Type) ([]ListedResource, []Diagnostic, error)
}

// ListedResource is one object found by ListResource.
type ListedResource struct {
	DisplayName string
	// Object is the full resource, or null when it was not requested or the
	// provider did not send it.
	Object cty.Value
}

// Provider is a running provider plugin.
type Provider struct {
	Path     string
	Protocol int

	client *goplugin.Client
	rpc    protocol

	schemaOnce sync.Once
	schema     *schema.Provider
	schemaErr  error
}

// Options tweak how the plugin process is started.
type Options struct {
	// Logger receives the plugin's log output. Nil keeps it quiet.
	Logger hclog.Logger
}

// Start launches the provider binary at path and negotiates a protocol.
func Start(path string, opts Options) (*Provider, error) {
	logger := opts.Logger
	if logger == nil {
		logger = hclog.New(&hclog.LoggerOptions{Name: "plugin", Level: hclog.Error})
	}
	client := goplugin.NewClient(&goplugin.ClientConfig{
		Cmd:              exec.Command(path),
		HandshakeConfig:  handshake,
		VersionedPlugins: versionedPlugins,
		AllowedProtocols: []goplugin.Protocol{goplugin.ProtocolGRPC},
		AutoMTLS:         true,
		Managed:          true,
		Logger:           logger,
		GRPCDialOptions: []grpc.DialOption{
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxRecvMsgSize)),
		},
	})
	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("starting provider %s: %w", path, err)
	}
	raw, err := rpcClient.Dispense("provider")
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("connecting to provider %s: %w", path, err)
	}
	return &Provider{
		Path:     path,
		Protocol: client.NegotiatedVersion(),
		client:   client,
		rpc:      raw.(protocol),
	}, nil
}

// Close stops the plugin process.
func (p *Provider) Close() {
	if p != nil && p.client != nil {
		p.client.Kill()
	}
}

// Schema returns the provider's schema, fetching it on first use.
func (p *Provider) Schema(ctx context.Context) (*schema.Provider, error) {
	p.schemaOnce.Do(func() {
		p.schema, p.schemaErr = p.rpc.getSchema(ctx)
	})
	return p.schema, p.schemaErr
}

// Configure passes the provider configuration. Attributes missing from
// config are sent as null, the same as when they are omitted in HCL.
func (p *Provider) Configure(ctx context.Context, config cty.Value) error {
	s, err := p.Schema(ctx)
	if err != nil {
		return err
	}
	coerced, err := s.Provider.CoerceValue(config)
	if err != nil {
		return fmt.Errorf("provider configuration: %w", err)
	}
	return p.rpc.configure(ctx, coerced, s.Provider.ImpliedType())
}

// ReadResource refreshes an object from the remote API. A null result means
// the object no longer exists.
func (p *Provider) ReadResource(ctx context.Context, typeName string, state cty.Value, private []byte) (cty.Value, []byte, error) {
	ty, err := p.resourceType(ctx, typeName)
	if err != nil {
		return cty.NilVal, nil, err
	}
	return p.rpc.readResource(ctx, typeName, state, ty, private)
}

// ImportResourceState asks the provider to adopt the object with the given
// import ID, exactly as an `import` block would.
func (p *Provider) ImportResourceState(ctx context.Context, typeName, id string) ([]ImportedResource, error) {
	s, err := p.Schema(ctx)
	if err != nil {
		return nil, err
	}
	return p.rpc.importResourceState(ctx, typeName, id, func(name string) (cty.Type, bool) {
		r, ok := s.ResourceTypes[name]
		if !ok {
			return cty.NilType, false
		}
		return r.Block.ImpliedType(), true
	})
}

// ValidateResourceConfig checks a resource configuration the way `tofu
// validate` does, including rules the schema cannot express, such as
// arguments that conflict with or require each other. config must conform
// to the resource type's schema; unknown values stand for references.
func (p *Provider) ValidateResourceConfig(ctx context.Context, typeName string, config cty.Value) ([]Diagnostic, error) {
	ty, err := p.resourceType(ctx, typeName)
	if err != nil {
		return nil, err
	}
	return p.rpc.validateResourceConfig(ctx, typeName, config, ty)
}

// ListResource asks the provider for existing objects of a resource type
// that has a list resource, the mechanism behind Terraform's `query`
// command. config follows the list resource's own schema (filters and the
// like); missing attributes are null. Objects whose event carried an error
// are skipped; all diagnostics are returned.
func (p *Provider) ListResource(ctx context.Context, typeName string, config cty.Value, includeResource bool, limit int64) ([]ListedResource, []Diagnostic, error) {
	s, err := p.Schema(ctx)
	if err != nil {
		return nil, nil, err
	}
	lr, ok := s.ListResources[typeName]
	if !ok {
		return nil, nil, fmt.Errorf("provider has no list resource for %q", typeName)
	}
	rs, ok := s.ResourceTypes[typeName]
	if !ok {
		return nil, nil, fmt.Errorf("provider does not support resource type %q", typeName)
	}
	coerced, err := lr.Block.CoerceValue(config)
	if err != nil {
		return nil, nil, fmt.Errorf("%s list configuration: %w", typeName, err)
	}
	return p.rpc.listResource(ctx, typeName, coerced, lr.Block.ImpliedType(), includeResource, limit, rs.Block.ImpliedType())
}

func (p *Provider) resourceType(ctx context.Context, typeName string) (cty.Type, error) {
	s, err := p.Schema(ctx)
	if err != nil {
		return cty.NilType, err
	}
	r, ok := s.ResourceTypes[typeName]
	if !ok {
		return cty.NilType, fmt.Errorf("provider does not support resource type %q", typeName)
	}
	return r.Block.ImpliedType(), nil
}
