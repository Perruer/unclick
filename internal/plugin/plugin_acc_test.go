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

package plugin_test

import (
	"context"
	"os"
	"testing"

	"github.com/zclconf/go-cty/cty"

	"github.com/Perruer/unclick/internal/install"
	"github.com/Perruer/unclick/internal/plugin"
	"github.com/Perruer/unclick/internal/shim"
)

// startProvider downloads (once) and starts a real provider. It needs tofu
// or terraform in PATH and network access, so it only runs with
// UNCLICK_ACC=1.
func startProvider(t *testing.T, source string) *plugin.Provider {
	t.Helper()
	if os.Getenv("UNCLICK_ACC") == "" {
		t.Skip("set UNCLICK_ACC=1 to run tests against real provider plugins")
	}
	src, err := install.ParseSource(source)
	if err != nil {
		t.Fatal(err)
	}
	cache := os.Getenv("UNCLICK_CACHE_DIR")
	if cache == "" {
		cache = t.TempDir()
	}
	found, err := (&install.Installer{CacheDir: cache}).Ensure(context.Background(), src, "")
	if err != nil {
		t.Fatal(err)
	}
	p, err := plugin.Start(found.Path, plugin.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(p.Close)
	return p
}

func TestAccImportAndRead(t *testing.T) {
	ctx := context.Background()
	p := startProvider(t, "hashicorp/random")
	s, err := p.Schema(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s.ResourceTypes["random_string"]; !ok {
		t.Fatal("random provider has no random_string resource")
	}
	if err := p.Configure(ctx, cty.NilVal); err != nil {
		t.Fatal(err)
	}

	imported, err := p.ImportResourceState(ctx, "random_string", "unclick")
	if err != nil {
		t.Fatal(err)
	}
	if len(imported) != 1 {
		t.Fatalf("import returned %d objects", len(imported))
	}
	state, _, err := p.ReadResource(ctx, imported[0].TypeName, imported[0].State, imported[0].Private)
	if err != nil {
		t.Fatal(err)
	}
	if got := state.GetAttr("result").AsString(); got != "unclick" {
		t.Fatalf("result = %q", got)
	}

	// The importers still pass resources around as flatmaps, so the round
	// trip through that format must be lossless.
	back, err := shim.HCL2ValueFromFlatmap(shim.FlatmapValueFromHCL2(state), s.ResourceTypes["random_string"].Block.ImpliedType())
	if err != nil {
		t.Fatal(err)
	}
	if !back.RawEquals(state) {
		t.Fatalf("flatmap round trip changed the value:\n%#v\n%#v", state, back)
	}
}

func TestAccUnknownResourceType(t *testing.T) {
	p := startProvider(t, "hashicorp/random")
	if _, _, err := p.ReadResource(context.Background(), "random_nope", cty.NilVal, nil); err == nil {
		t.Fatal("expected an error for an unknown resource type")
	}
}
