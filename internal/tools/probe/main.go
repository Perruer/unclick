// Command probe starts a real provider plugin with the new client and
// prints what it reports. It is a development aid:
//
//	go run ./internal/tools/probe hashicorp/random random_string some-id
//
// PROBE_CONFIG_JSON holds the provider configuration as a JSON object and
// PROBE_SKIP_CONFIGURE=1 only fetches the schema, and PROBE_LIST=<type> calls
// ListResource for that type.
package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"

	"github.com/Perruer/unclick/internal/install"
	"github.com/Perruer/unclick/internal/plugin"
	"github.com/Perruer/unclick/internal/shim"
)

func main() {
	ctx := context.Background()
	src, _ := install.ParseSource(os.Args[1])
	in := &install.Installer{CacheDir: os.Getenv("UNCLICK_CACHE_DIR")}
	found, err := in.Ensure(ctx, src, "")
	check(err)
	fmt.Println("provider:", found.Source, found.Version, found.Path)
	p, err := plugin.Start(found.Path, plugin.Options{})
	check(err)
	defer p.Close()
	s, err := p.Schema(ctx)
	check(err)
	fmt.Printf("protocol v%d, %d resource types, %d data sources, %d list resources\n",
		p.Protocol, len(s.ResourceTypes), len(s.DataSources), len(s.ListResources))
	var cfg cty.Value
	if js := os.Getenv("PROBE_CONFIG_JSON"); js != "" {
		ty, err := ctyjson.ImpliedType([]byte(js))
		check(err)
		cfg, err = ctyjson.Unmarshal([]byte(js), ty)
		check(err)
	}
	if os.Getenv("PROBE_SKIP_CONFIGURE") == "" {
		check(p.Configure(ctx, cfg))
	}
	if typ := os.Getenv("PROBE_LIST"); typ != "" {
		found, diags, err := p.ListResource(ctx, typ, cty.NilVal, true, 1000)
		check(err)
		for _, d := range diags {
			fmt.Printf("diagnostic: %s: %s\n", d.Summary, d.Detail)
		}
		fmt.Printf("list %s -> %d object(s)\n", typ, len(found))
		for _, f := range found {
			id := "?"
			if !f.Object.IsNull() && f.Object.Type().HasAttribute("id") && !f.Object.GetAttr("id").IsNull() {
				id = f.Object.GetAttr("id").AsString()
			}
			fmt.Printf("  %q id=%s attributes=%d\n", f.DisplayName, id, len(shim.FlatmapValueFromHCL2(f.Object)))
		}
		return
	}
	if len(os.Args) > 3 {
		typ, id := os.Args[2], os.Args[3]
		imported, err := p.ImportResourceState(ctx, typ, id)
		check(err)
		fmt.Printf("import %s %q -> %d object(s)\n", typ, id, len(imported))
		st, _, err := p.ReadResource(ctx, imported[0].TypeName, imported[0].State, imported[0].Private)
		check(err)
		flat := shim.FlatmapValueFromHCL2(st)
		keys := make([]string, 0, len(flat))
		for k := range flat {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		fmt.Printf("read back %d flatmap attributes:\n", len(flat))
		for i, k := range keys {
			if i >= 12 {
				fmt.Println("  ...")
				break
			}
			fmt.Printf("  %s = %q\n", k, flat[k])
		}
		back, err := shim.HCL2ValueFromFlatmap(flat, s.ResourceTypes[typ].Block.ImpliedType())
		check(err)
		fmt.Println("flatmap round trip equal:", back.Equals(st).True())
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}
