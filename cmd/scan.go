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

package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/spf13/cobra"
	"github.com/zclconf/go-cty/cty"

	"github.com/Perruer/unclick/internal/install"
	"github.com/Perruer/unclick/terraformutils"
	"github.com/Perruer/unclick/terraformutils/providerwrapper"
	"github.com/Perruer/unclick/terraformutils/terraformoutput"
)

type scanOptions struct {
	Provider     string
	Config       []string
	Types        []string
	ExcludeTypes []string
	PathOutput   string
	Limit        int64
	Verbose      bool
}

func newScanCmd() *cobra.Command {
	options := scanOptions{}
	cmd := &cobra.Command{
		Use:   "scan PROVIDER",
		Short: "Find existing resources through the provider's list resources and write them as code",
		Long: `Scan asks the provider itself for every existing object of every resource
type it can list, then writes configuration and import blocks for them.

It needs no Unclick code per resource type: whatever the provider can list
is covered, and new provider releases add coverage by themselves. Providers
gained list resources with Terraform 1.14 (AWS, Google and AzureRM so far);
OpenTofu cannot query them yet, Unclick can.

PROVIDER is a name such as "aws" or a registry address such as
"hashicorp/aws". Provider settings go in --config; credentials come from the
environment, exactly as for OpenTofu.`,
		Example: `  unclick scan aws --config region=eu-west-1
  unclick scan aws --config region=us-east-1 --types aws_vpc,aws_subnet
  unclick scan hashicorp/google --config project=my-project --config region=europe-west1`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.Provider = args[0]
			return runScan(options)
		},
	}
	f := cmd.Flags()
	f.StringArrayVarP(&options.Config, "config", "c", nil, "provider setting as key=value; repeat for several")
	f.StringSliceVarP(&options.Types, "types", "t", nil, "resource types to scan (default: every type with a list resource)")
	f.StringSliceVarP(&options.ExcludeTypes, "exclude-types", "x", nil, "resource types to skip")
	f.StringVarP(&options.PathOutput, "path-output", "o", DefaultPathOutput, "directory to write into; files go to <path-output>/<provider>/")
	f.Int64Var(&options.Limit, "limit", 10000, "maximum number of objects per resource type")
	f.BoolVarP(&options.Verbose, "verbose", "v", false, "log provider plugin output")
	f.StringVar(&providerwrapper.TFBinary, "tf-binary", providerwrapper.TFBinary, "tofu or terraform executable used to download providers (default: tofu, then terraform, from PATH)")
	f.StringVar(&providerwrapper.ProviderVersion, "provider-version", providerwrapper.ProviderVersion, "version constraint for the provider plugin, e.g. \"~> 6.0\"")
	return cmd
}

func runScan(options scanOptions) error {
	name := options.Provider
	if strings.Contains(name, "/") {
		src, err := install.ParseSource(name)
		if err != nil {
			return err
		}
		name = src.Type
		install.KnownSources[name] = src.String()
	}
	configValues := map[string]string{}
	config := map[string]cty.Value{}
	for _, pair := range options.Config {
		k, v, ok := strings.Cut(pair, "=")
		if !ok || k == "" {
			return fmt.Errorf("--config %q: expected key=value", pair)
		}
		configValues[k] = v
		config[k] = cty.StringVal(v)
	}

	pw, err := providerwrapper.NewProviderWrapper(name, cty.ObjectVal(config), options.Verbose)
	if err != nil {
		return err
	}
	defer pw.Kill()

	types, err := scanTypes(pw, options)
	if err != nil {
		return err
	}
	resources, err := listResources(pw, name, types, options.Limit)
	if err != nil {
		return err
	}
	if len(resources) == 0 {
		log.Printf("no %s resources found", name)
		return nil
	}

	usedTypes := map[string]bool{}
	for _, r := range resources {
		usedTypes[r.InstanceInfo.Type] = true
	}
	var typeList []string
	for t := range usedTypes {
		typeList = append(typeList, t)
	}
	readOnly, _ := pw.GetReadOnlyAttributes(typeList)
	zeros := pw.GetZeroNumberAttributes(typeList)
	for _, r := range resources {
		r.IgnoreKeys = append(r.IgnoreKeys, readOnly[r.InstanceInfo.Type]...)
		r.ZeroValueKeys = append(r.ZeroValueKeys, zeros[r.InstanceInfo.Type]...)
		if err := r.ConvertTFstate(pw); err != nil {
			log.Printf("failed to convert %s: %v", r.InstanceInfo.Id, err)
		}
	}
	terraformutils.LinkByID(resources)
	terraformutils.LogFixSummary(terraformutils.FixInvalidConfig(resources, pw))

	dir := filepath.Join(options.PathOutput, name)
	if err := writeScanOutput(dir, name, configValues, resources); err != nil {
		return err
	}
	log.Printf("wrote %d %s resources of %d types to %s", len(resources), name, len(typeList), dir)
	return nil
}

func scanTypes(pw *providerwrapper.ProviderWrapper, options scanOptions) ([]string, error) {
	listable := pw.GetSchema().ListResources
	if len(listable) == 0 {
		return nil, fmt.Errorf("provider %s has no list resources; use `unclick import %s` instead", options.Provider, options.Provider)
	}
	excluded := map[string]bool{}
	for _, t := range options.ExcludeTypes {
		excluded[t] = true
	}
	var types []string
	if len(options.Types) > 0 {
		for _, t := range options.Types {
			if _, ok := listable[t]; !ok {
				return nil, fmt.Errorf("provider %s cannot list %s", options.Provider, t)
			}
			types = append(types, t)
		}
	} else {
		for t := range listable {
			types = append(types, t)
		}
	}
	out := types[:0]
	for _, t := range types {
		if !excluded[t] {
			out = append(out, t)
		}
	}
	sort.Strings(out)
	return out, nil
}

func listResources(pw *providerwrapper.ProviderWrapper, provider string, types []string, limit int64) ([]*terraformutils.Resource, error) {
	schema := pw.GetSchema()
	var resources []*terraformutils.Resource
	for _, t := range types {
		found, diags, err := pw.ListResource(t, limit)
		for _, d := range diags {
			if d.Error {
				log.Printf("WARN: %s: %s: %s", t, d.Summary, d.Detail)
			}
		}
		if err != nil {
			log.Printf("WARN: cannot list %s: %v", t, err)
			continue
		}
		names := map[string]int{}
		for _, f := range found {
			if f.Object.IsNull() {
				log.Printf("WARN: %s %q: the provider did not return the object", t, f.DisplayName)
				continue
			}
			state := providerwrapper.StateFromValue(f.Object, schema.ResourceTypes[t].Version)
			if state.ID == "" {
				log.Printf("WARN: %s %q has no id and cannot be imported", t, f.DisplayName)
				continue
			}
			resourceName := state.ID
			if n := names[resourceName]; n > 0 {
				resourceName = fmt.Sprintf("%s_%d", resourceName, n+1)
			}
			names[state.ID]++
			r := terraformutils.NewResource(state.ID, resourceName, t, provider, state.Attributes, nil, nil)
			r.InstanceState = state
			resources = append(resources, &r)
		}
		if len(found) > 0 {
			log.Printf("%s: %d found", t, len(found))
		}
	}
	return resources, nil
}

func writeScanOutput(dir, provider string, config map[string]string, resources []*terraformutils.Resource) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "provider.tf"), providerFile(provider, config), 0o644); err != nil {
		return err
	}
	byType := map[string][]terraformutils.Resource{}
	var all []terraformutils.Resource
	for _, r := range resources {
		byType[r.InstanceInfo.Type] = append(byType[r.InstanceInfo.Type], *r)
		all = append(all, *r)
	}
	for t, rs := range byType {
		body, err := terraformutils.WriteResourcesHCL(rs)
		if err != nil {
			return fmt.Errorf("%s: %w", t, err)
		}
		file := strings.TrimPrefix(t, provider+"_") + ".tf"
		if err := os.WriteFile(filepath.Join(dir, file), body, 0o644); err != nil {
			return err
		}
	}
	return terraformoutput.OutputImportFile(all, dir)
}

// providerFile renders required_providers with the release that was used,
// and the provider block with the settings given on the command line.
func providerFile(provider string, config map[string]string) []byte {
	f := hclwrite.NewEmptyFile()
	req := f.Body().AppendNewBlock("terraform", nil).Body().AppendNewBlock("required_providers", nil).Body()
	entry := map[string]cty.Value{"source": cty.StringVal(providerwrapper.GetProviderSource(provider))}
	if v := providerwrapper.GetProviderVersion(provider); v != "" {
		entry["version"] = cty.StringVal(v)
	}
	req.SetAttributeValue(provider, cty.ObjectVal(entry))
	f.Body().AppendNewline()
	block := f.Body().AppendNewBlock("provider", []string{provider}).Body()
	keys := make([]string, 0, len(config))
	for k := range config {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		block.SetAttributeValue(k, cty.StringVal(config[k]))
	}
	return f.Bytes()
}
