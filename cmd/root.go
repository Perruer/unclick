// Copyright 2018 The Terraformer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/Perruer/unclick/terraformutils"
	"github.com/spf13/cobra"
)

func NewCmdRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "unclick",
		Short:         "Turn existing cloud resources into OpenTofu / Terraform code",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}
	cmd.AddCommand(newImportCmd())
	cmd.AddCommand(newPlanCmd())
	cmd.AddCommand(newScanCmd())
	cmd.AddCommand(versionCmd)
	return cmd
}

func Execute() error {
	cmd := NewCmdRoot()
	return cmd.Execute()
}

// registeredProvider is one provider compiled into this binary. Each
// provider_cmd_*.go file registers itself, so a build with `-tags slim,aws`
// contains only the AWS importer.
type registeredProvider struct {
	importer  func(options ImportOptions) *cobra.Command
	generator func() terraformutils.ProviderGenerator
}

var registeredProviders []registeredProvider

func registerProvider(importer func(options ImportOptions) *cobra.Command, generator func() terraformutils.ProviderGenerator) {
	registeredProviders = append(registeredProviders, registeredProvider{importer: importer, generator: generator})
}

func providerImporterSubcommands() []func(options ImportOptions) *cobra.Command {
	list := make([]func(options ImportOptions) *cobra.Command, 0, len(registeredProviders))
	for _, p := range registeredProviders {
		list = append(list, p.importer)
	}
	return list
}

func providerGenerators() map[string]func() terraformutils.ProviderGenerator {
	list := make(map[string]func() terraformutils.ProviderGenerator, len(registeredProviders))
	for _, p := range registeredProviders {
		list[p.generator().GetName()] = p.generator
	}
	return list
}
