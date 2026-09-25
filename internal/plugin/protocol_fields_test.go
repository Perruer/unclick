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
	"sort"
	"strings"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"

	p5 "github.com/Perruer/unclick/internal/tfplugin5"
	p6 "github.com/Perruer/unclick/internal/tfplugin6"
)

// fieldDecisions says, for every protocol message the adapters read, which
// fields they carry over and which they leave out on purpose. The field
// lists themselves come from the generated descriptors, so when a newer
// tfplugin5.proto or tfplugin6.proto adds a field, this test fails until
// someone decides what the adapters should do with it.
//
// Messages are keyed by their name inside the package, so one entry covers
// both protocol versions; fields that exist in only one version (nested_type
// in 6) are fine, but a field present in both gets the same decision.
var fieldDecisions = map[string]struct {
	mapped  []string
	skipped map[string]string // field -> why it is left out
}{
	"Schema": {mapped: []string{"version", "block"}},
	"Schema.Block": {
		mapped: []string{"attributes", "block_types", "deprecated"},
		skipped: map[string]string{
			"version":             "the resource version is taken from Schema.version",
			"description":         "documentation only",
			"description_kind":    "documentation only",
			"deprecation_message": "documentation only; the deprecated flag is kept",
		},
	},
	"Schema.Attribute": {
		mapped: []string{"name", "type", "nested_type", "required", "optional", "computed", "sensitive", "deprecated", "write_only"},
		skipped: map[string]string{
			"description":         "documentation only",
			"description_kind":    "documentation only",
			"deprecation_message": "documentation only; the deprecated flag is kept",
		},
	},
	"Schema.NestedBlock": {mapped: []string{"type_name", "block", "nesting", "min_items", "max_items"}},
	"Schema.Object": {
		mapped: []string{"attributes", "nesting"},
		skipped: map[string]string{
			"min_items": "deprecated in protocol 6 and not sent by providers",
			"max_items": "deprecated in protocol 6 and not sent by providers",
		},
	},
	"Diagnostic":         {mapped: []string{"severity", "summary", "detail", "attribute"}},
	"AttributePath.Step": {mapped: []string{"attribute_name", "element_key_string", "element_key_int"}},
	"GetProviderSchema.Response": {
		mapped: []string{"provider", "resource_schemas", "data_source_schemas", "list_resource_schemas", "diagnostics"},
		skipped: map[string]string{
			"provider_meta":              "Unclick sends no provider_meta",
			"server_capabilities":        "optional optimisations Unclick does not use",
			"functions":                  "provider functions are not generated",
			"ephemeral_resource_schemas": "ephemeral resources have nothing to import",
			"action_schemas":             "actions have nothing to import",
			"state_store_schemas":        "state storage is not Unclick's business",
		},
	},
	"ReadResource.Response": {
		mapped: []string{"new_state", "diagnostics", "private"},
		skipped: map[string]string{
			"deferred":     "the client does not announce deferral_allowed, so providers must not defer",
			"new_identity": "imports are written with IDs, not resource identities",
		},
	},
	"ImportResourceState.Response": {
		mapped:  []string{"imported_resources", "diagnostics"},
		skipped: map[string]string{"deferred": "the client does not announce deferral_allowed"},
	},
	"ImportResourceState.ImportedResource": {
		mapped:  []string{"type_name", "state", "private"},
		skipped: map[string]string{"identity": "imports are written with IDs, not resource identities"},
	},
	"ValidateResourceTypeConfig.Response": {mapped: []string{"diagnostics"}},
	"ValidateResourceConfig.Response":     {mapped: []string{"diagnostics"}},
	"ListResource.Event": {
		mapped:  []string{"display_name", "resource_object", "diagnostic"},
		skipped: map[string]string{"identity": "scan reads resource objects, not identities"},
	},
	"PrepareProviderConfig.Response": {mapped: []string{"prepared_config", "diagnostics"}},
	"Configure.Response":             {mapped: []string{"diagnostics"}},
	"ConfigureProvider.Response":     {mapped: []string{"diagnostics"}},
}

// readMessages are the messages the v5 and v6 adapters read from providers.
var readMessages = []protoreflect.ProtoMessage{
	&p5.Schema{}, &p5.Schema_Block{}, &p5.Schema_Attribute{}, &p5.Schema_NestedBlock{},
	&p5.Diagnostic{}, &p5.AttributePath_Step{},
	&p5.GetProviderSchema_Response{}, &p5.ReadResource_Response{},
	&p5.ImportResourceState_Response{}, &p5.ImportResourceState_ImportedResource{},
	&p5.ValidateResourceTypeConfig_Response{}, &p5.ListResource_Event{},
	&p5.PrepareProviderConfig_Response{}, &p5.Configure_Response{},

	&p6.Schema{}, &p6.Schema_Block{}, &p6.Schema_Attribute{}, &p6.Schema_NestedBlock{}, &p6.Schema_Object{},
	&p6.Diagnostic{}, &p6.AttributePath_Step{},
	&p6.GetProviderSchema_Response{}, &p6.ReadResource_Response{},
	&p6.ImportResourceState_Response{}, &p6.ImportResourceState_ImportedResource{},
	&p6.ValidateResourceConfig_Response{}, &p6.ListResource_Event{},
	&p6.ConfigureProvider_Response{},
}

func TestAdaptersDecideOnEveryProtocolField(t *testing.T) {
	seen := map[string]map[string]bool{} // message -> fields present in some version
	for _, m := range readMessages {
		md := m.ProtoReflect().Descriptor()
		full := string(md.FullName())
		name := full[strings.Index(full, ".")+1:] // drop "tfplugin5."/"tfplugin6."
		decision, ok := fieldDecisions[name]
		if !ok {
			t.Errorf("%s: no field decisions", full)
			continue
		}
		known := map[string]bool{}
		for _, f := range decision.mapped {
			known[f] = true
		}
		for f := range decision.skipped {
			if known[f] {
				t.Errorf("%s: %s is both mapped and skipped", name, f)
			}
			known[f] = true
		}
		if seen[name] == nil {
			seen[name] = map[string]bool{}
		}
		fields := md.Fields()
		for i := 0; i < fields.Len(); i++ {
			f := string(fields.Get(i).Name())
			seen[name][f] = true
			if !known[f] {
				t.Errorf("%s has field %q (number %d) that the adapters neither map nor skip on purpose; "+
					"map it in v5.go/v6.go or add it to skipped with a reason", full, f, fields.Get(i).Number())
			}
		}
	}

	// A decision about a field no protocol version has any more is stale.
	for name, decision := range fieldDecisions {
		var stale []string
		for _, f := range decision.mapped {
			if !seen[name][f] {
				stale = append(stale, f)
			}
		}
		for f := range decision.skipped {
			if !seen[name][f] {
				stale = append(stale, f)
			}
		}
		sort.Strings(stale)
		if len(stale) > 0 {
			t.Errorf("%s: decisions for fields no protocol version has: %v", name, stale)
		}
	}
}
