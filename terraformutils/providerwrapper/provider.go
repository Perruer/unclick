// Copyright 2018 The Terraformer Authors.
// Copyright 2026 Perruer (Unclick modifications)
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

// Package providerwrapper runs one provider plugin for an import and hides
// the plugin protocol behind the flatmap-based API the importers expect.
package providerwrapper

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	version "github.com/hashicorp/go-version"
	"github.com/zclconf/go-cty/cty"

	"github.com/Perruer/unclick/internal/install"
	"github.com/Perruer/unclick/internal/legacy"
	"github.com/Perruer/unclick/internal/plugin"
	"github.com/Perruer/unclick/internal/schema"
	"github.com/Perruer/unclick/internal/shim"
	"github.com/Perruer/unclick/terraformutils/terraformerstring"
)

// Settings that the command line passes down before any provider starts.
var (
	// TFBinary is the tofu or terraform executable used to download
	// providers. Empty means look in PATH.
	TFBinary = os.Getenv("UNCLICK_TF_BINARY")
	// ProviderVersion constrains the provider release to download, for
	// example "~> 6.0". Empty means the newest one already installed, or the
	// newest one in the registry.
	ProviderVersion = ""
)

var (
	startedMu sync.Mutex
	started   = map[string]*install.Provider{}
)

type ProviderWrapper struct {
	provider     *plugin.Provider
	providerName string
	config       cty.Value
	retryCount   int
	retrySleepMs int
}

func NewProviderWrapper(providerName string, providerConfig cty.Value, verbose bool, options ...map[string]int) (*ProviderWrapper, error) {
	p := &ProviderWrapper{retryCount: 5, retrySleepMs: 300}
	p.providerName = providerName
	p.config = providerConfig

	if len(options) > 0 {
		retryCount, hasOption := options[0]["retryCount"]
		if hasOption {
			p.retryCount = retryCount
		}
		retrySleepMs, hasOption := options[0]["retrySleepMs"]
		if hasOption {
			p.retrySleepMs = retrySleepMs
		}
	}

	err := p.initProvider(verbose)

	return p, err
}

func (p *ProviderWrapper) Kill() {
	p.provider.Close()
}

// GetSchema returns the provider schema. It panics if the provider cannot
// report one, because nothing else works without it and initProvider has
// already fetched it successfully.
func (p *ProviderWrapper) GetSchema() *schema.Provider {
	s, err := p.provider.Schema(context.Background())
	if err != nil {
		panic(err)
	}
	return s
}

func (p *ProviderWrapper) GetReadOnlyAttributes(resourceTypes []string) (map[string][]string, error) {
	return readOnlyAttributesOf(p.GetSchema(), resourceTypes), nil
}

// readOnlyAttributesOf returns, per resource type, patterns for the flatmap
// keys that must not be written to configuration.
func readOnlyAttributesOf(r *schema.Provider, resourceTypes []string) map[string][]string {
	readOnlyAttributes := map[string][]string{}
	for resourceName, obj := range r.ResourceTypes {
		if terraformerstring.ContainsString(resourceTypes, resourceName) {
			readOnlyAttributes[resourceName] = append(readOnlyAttributes[resourceName], "^id$")
			for k, v := range obj.Block.Attributes {
				// Deprecated arguments are left out too: they usually have a
				// replacement that is also set, and writing both fails with
				// "conflicts with" errors (github_repository.private vs
				// visibility, for example).
				if (!v.Optional && !v.Required) || v.Deprecated {
					ty := v.ImpliedType()
					if ty.IsListType() || ty.IsSetType() {
						readOnlyAttributes[resourceName] = append(readOnlyAttributes[resourceName], "^"+k+"\\.(.*)")
					} else {
						readOnlyAttributes[resourceName] = append(readOnlyAttributes[resourceName], "^"+k+"$")
					}
				}
			}
			for k, v := range obj.Block.BlockTypes {
				if v.Deprecated {
					readOnlyAttributes[resourceName] = append(readOnlyAttributes[resourceName], "^"+k+`($|\.)`)
				}
			}
			readOnlyAttributes[resourceName] = readObjBlocks(obj.Block.BlockTypes, readOnlyAttributes[resourceName], "-1")
		}
	}
	return readOnlyAttributes
}

// ListResource returns the existing objects of a resource type that has a
// list resource, with their full state.
func (p *ProviderWrapper) ListResource(typeName string, limit int64) ([]plugin.ListedResource, []plugin.Diagnostic, error) {
	return p.provider.ListResource(context.Background(), typeName, cty.NilVal, true, limit)
}

// ValidateResourceConfig asks the provider to validate a resource body.
func (p *ProviderWrapper) ValidateResourceConfig(typeName string, config cty.Value) ([]plugin.Diagnostic, error) {
	return p.provider.ValidateResourceConfig(context.Background(), typeName, config)
}

// GetZeroNumberAttributes returns, per resource type, patterns for optional
// number attributes whose value 0 must not be written to configuration.
func (p *ProviderWrapper) GetZeroNumberAttributes(resourceTypes []string) map[string][]string {
	return zeroNumberAttributesOf(p.GetSchema(), resourceTypes)
}

// zeroNumberAttributesOf covers resources built with the legacy plugin SDK
// (SDKv2), which stores 0 rather than null for a number that was never set.
// Writing that 0 back can break rules the schema does not expose, such as
// aws_vpc.ipv6_netmask_length needing ipv6_ipam_pool_id; leaving it out reads
// back as 0, so the plan does not change. Booleans are kept because a false
// is often an explicit override of a true default.
func zeroNumberAttributesOf(r *schema.Provider, resourceTypes []string) map[string][]string {
	out := map[string][]string{}
	for _, name := range resourceTypes {
		rs, ok := r.ResourceTypes[name]
		if !ok || !isLegacySDKResource(rs.Block) {
			continue
		}
		if patterns := zeroNumberPatterns(rs.Block, "^"); len(patterns) > 0 {
			out[name] = patterns
		}
	}
	return out
}

// isLegacySDKResource recognizes SDKv2 resources by the "id" attribute the
// SDK adds to every resource as optional and computed. Plugin framework
// resources declare id themselves, as computed only.
func isLegacySDKResource(b *schema.Block) bool {
	id := b.Attributes["id"]
	return id != nil && id.Optional && id.Computed
}

func zeroNumberPatterns(b *schema.Block, prefix string) []string {
	var out []string
	for name, a := range b.Attributes {
		if a.Optional && !a.Required && a.ImpliedType().Equals(cty.Number) {
			out = append(out, prefix+regexp.QuoteMeta(name)+"$")
		}
	}
	for name, nb := range b.BlockTypes {
		// SDKv2 flattens every nested block as a list or a set.
		out = append(out, zeroNumberPatterns(&nb.Block, prefix+regexp.QuoteMeta(name)+`\.[0-9]+\.`)...)
	}
	return out
}

func readObjBlocks(block map[string]*schema.NestedBlock, readOnlyAttributes []string, parent string) []string {
	for k, v := range block {
		if len(v.BlockTypes) > 0 {
			if parent == "-1" {
				readOnlyAttributes = readObjBlocks(v.BlockTypes, readOnlyAttributes, k)
			} else {
				readOnlyAttributes = readObjBlocks(v.BlockTypes, readOnlyAttributes, parent+"\\.[0-9]+\\."+k)
			}
		}
		fieldCount := 0
		for key, l := range v.Attributes {
			if (!l.Optional && !l.Required) || l.Deprecated {
				fieldCount++
				switch v.Nesting {
				case schema.NestingList:
					if parent == "-1" {
						readOnlyAttributes = append(readOnlyAttributes, "^"+k+"\\.[0-9]+\\."+key+"($|\\.[0-9]+|\\.#)")
					} else {
						readOnlyAttributes = append(readOnlyAttributes, "^"+parent+"\\.(.*)\\."+key+"$")
					}
				case schema.NestingSet:
					if parent == "-1" {
						readOnlyAttributes = append(readOnlyAttributes, "^"+k+"\\.[0-9]+\\."+key+"$")
					} else {
						readOnlyAttributes = append(readOnlyAttributes, "^"+parent+"\\.(.*)\\."+key+"($|\\.(.*))")
					}
				case schema.NestingMap:
					readOnlyAttributes = append(readOnlyAttributes, parent+"\\."+key)
				default:
					readOnlyAttributes = append(readOnlyAttributes, parent+"\\."+key+"$")
				}
			}
		}
		if fieldCount == len(v.Block.Attributes) && fieldCount > 0 && len(v.BlockTypes) == 0 {
			readOnlyAttributes = append(readOnlyAttributes, "^"+k)
		}
	}
	return readOnlyAttributes
}

// Refresh reads the current attributes of a resource. Importers usually know
// only the ID and a few attributes, which is enough for the provider to look
// the object up. When reading fails, the object is imported by ID instead.
func (p *ProviderWrapper) Refresh(info *legacy.InstanceInfo, state *legacy.InstanceState) (*legacy.InstanceState, error) {
	ctx := context.Background()
	rs, ok := p.GetSchema().ResourceTypes[info.Type]
	if !ok {
		return nil, fmt.Errorf("provider %s does not support resource type %s", p.providerName, info.Type)
	}
	priorState, err := attrsAsObjectValue(state, rs.Block.ImpliedType())
	if err != nil {
		return nil, err
	}

	var newState cty.Value
	var readErr error
	for i := 0; i < p.retryCount; i++ {
		newState, _, readErr = p.provider.ReadResource(ctx, info.Type, priorState, nil)
		if readErr == nil {
			break
		}
		log.Println(readErr)
		log.Printf("WARN: Fail read resource from provider, wait %dms before retry\n", p.retrySleepMs)
		time.Sleep(time.Duration(p.retrySleepMs) * time.Millisecond)
	}

	if readErr != nil {
		log.Println("Fail read resource from provider, trying import command")
		// retry with regular import command - without resource attributes
		imported, err := p.provider.ImportResourceState(ctx, info.Type, state.ID)
		if err != nil {
			return nil, readErr
		}
		if len(imported) == 0 {
			return nil, errors.New("not able to import resource for a given ID")
		}
		// Imported state often holds little more than the ID, so read it
		// once more, the way `tofu import` does.
		newState, _, err = p.provider.ReadResource(ctx, imported[0].TypeName, imported[0].State, imported[0].Private)
		if err != nil {
			return nil, err
		}
	}

	if newState.IsNull() {
		return nil, fmt.Errorf("ERROR: Read resource response is null for resource %s", info.Id)
	}
	return StateFromValue(newState, rs.Version), nil
}

func (p *ProviderWrapper) initProvider(verbose bool) error {
	src := install.SourceFor(p.providerName)
	installer := &install.Installer{Binary: TFBinary, CacheDir: cacheDir()}
	found, err := installer.Ensure(context.Background(), src, ProviderVersion)
	if err != nil {
		return err
	}
	level := hclog.Error
	if verbose {
		level = hclog.Trace
	}
	p.provider, err = plugin.Start(found.Path, plugin.Options{
		Logger: hclog.New(&hclog.LoggerOptions{Name: "plugin", Level: level, Output: os.Stdout}),
	})
	if err != nil {
		return err
	}
	if _, err := p.provider.Schema(context.Background()); err != nil {
		p.provider.Close()
		return err
	}
	if p.config != cty.NilVal {
		if err := p.provider.Configure(context.Background(), p.config); err != nil {
			p.provider.Close()
			return err
		}
	}

	startedMu.Lock()
	started[p.providerName] = found
	startedMu.Unlock()
	return nil
}

// GetProviderVersion returns a version constraint that matches the provider
// release the configuration was generated with, e.g. "~> 6.14".
func GetProviderVersion(providerName string) string {
	startedMu.Lock()
	found := started[providerName]
	startedMu.Unlock()
	if found == nil {
		var err error
		if found, err = install.Find(install.SourceFor(providerName), install.SearchDirs()); err != nil {
			log.Printf("Can't find provider %s: %v", providerName, err)
			return ""
		}
	}
	v, err := version.NewVersion(found.Version)
	if err != nil {
		return "~> " + found.Version
	}
	segs := v.Segments()
	if len(segs) < 2 {
		return "~> " + found.Version
	}
	return fmt.Sprintf("~> %d.%d", segs[0], segs[1])
}

// GetProviderSource returns the registry address to put in required_providers.
func GetProviderSource(providerName string) string {
	src := install.SourceFor(providerName)
	if src.Host == "registry.opentofu.org" {
		src.Host = ""
	}
	return src.String()
}

func cacheDir() string {
	if d := os.Getenv("UNCLICK_CACHE_DIR"); d != "" {
		return filepath.Join(d, "providers")
	}
	if d, err := os.UserCacheDir(); err == nil {
		return filepath.Join(d, "unclick", "providers")
	}
	return filepath.Join(os.TempDir(), "unclick", "providers")
}

// attrsAsObjectValue mirrors Terraform 0.12's InstanceState.AttrsAsObjectValue:
// the ID wins over an "id" attribute, and the flatmap is decoded by type.
func attrsAsObjectValue(s *legacy.InstanceState, ty cty.Type) (cty.Value, error) {
	if s == nil {
		return cty.NullVal(ty), nil
	}
	attrs := make(map[string]string, len(s.Attributes)+1)
	for k, v := range s.Attributes {
		attrs[k] = v
	}
	if s.ID != "" {
		attrs["id"] = s.ID
	}
	return shim.HCL2ValueFromFlatmap(attrs, ty)
}

// StateFromValue converts a resource object into the flatmap state the
// importers work with.
func StateFromValue(v cty.Value, schemaVersion int64) *legacy.InstanceState {
	attrs := shim.FlatmapValueFromHCL2(v)
	return &legacy.InstanceState{
		ID:         attrs["id"],
		Attributes: attrs,
		Meta:       map[string]interface{}{"schema_version": int(schemaVersion)},
	}
}
