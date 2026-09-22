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

package install

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestParseSource(t *testing.T) {
	cases := map[string]Source{
		"aws":             {Namespace: "hashicorp", Type: "aws"},
		"DataDog/datadog": {Namespace: "DataDog", Type: "datadog"},
		"registry.terraform.io/yandex-cloud/yandex": {Host: "registry.terraform.io", Namespace: "yandex-cloud", Type: "yandex"},
	}
	for in, want := range cases {
		got, err := ParseSource(in)
		if err != nil || got != want {
			t.Errorf("ParseSource(%q) = %+v, %v; want %+v", in, got, err, want)
		}
	}
	if _, err := ParseSource("a/b/c/d"); err == nil {
		t.Error("expected an error for a four-part source")
	}
}

func TestSourceFor(t *testing.T) {
	if got := SourceFor("datadog").String(); got != "DataDog/datadog" {
		t.Errorf("datadog source = %q", got)
	}
	if got := SourceFor("somethingnew").String(); got != "hashicorp/somethingnew" {
		t.Errorf("unknown provider source = %q", got)
	}
}

// fakeProvider creates an empty file where OpenTofu would unpack a provider.
func fakeProvider(t *testing.T, root, host, ns, typ, version, file string) string {
	t.Helper()
	dir := filepath.Join(root, host, ns, typ, version, runtime.GOOS+"_"+runtime.GOARCH)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, file)
	if err := os.WriteFile(path, nil, 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestFindPicksNewestAcrossRegistries(t *testing.T) {
	root := t.TempDir()
	fakeProvider(t, root, "registry.terraform.io", "hashicorp", "aws", "5.99.0", "terraform-provider-aws_v5.99.0_x5")
	want := fakeProvider(t, root, "registry.opentofu.org", "hashicorp", "aws", "6.2.0", "terraform-provider-aws_v6.2.0_x5")
	fakeProvider(t, root, "registry.opentofu.org", "hashicorp", "aws", "6.10.0", "README.txt") // no binary

	got, err := Find(Source{Namespace: "hashicorp", Type: "aws"}, []string{root})
	if err != nil {
		t.Fatal(err)
	}
	if got.Path != want || got.Version != "6.2.0" {
		t.Fatalf("Find = %+v, want %s", got, want)
	}
}

func TestFindSkipsLookAlikes(t *testing.T) {
	root := t.TempDir()
	fakeProvider(t, root, "registry.opentofu.org", "hashicorp", "google", "7.0.0", "terraform-provider-google-beta_v7.0.0")
	if _, err := Find(Source{Namespace: "hashicorp", Type: "google"}, []string{root}); err == nil {
		t.Fatal("google-beta must not be taken for google")
	}
}

func TestFindHonoursHost(t *testing.T) {
	root := t.TempDir()
	fakeProvider(t, root, "registry.opentofu.org", "yandex-cloud", "yandex", "0.150.0", "terraform-provider-yandex_v0.150.0")
	src := Source{Host: "registry.terraform.io", Namespace: "yandex-cloud", Type: "yandex"}
	if _, err := Find(src, []string{root}); err == nil {
		t.Fatal("a provider from another registry must not match an explicit host")
	}
}

func TestRequiredProviders(t *testing.T) {
	got := requiredProviders(Source{Namespace: "DataDog", Type: "datadog"}, "~> 3.0")
	for _, want := range []string{`datadog = {`, `source = "DataDog/datadog"`, `version = "~> 3.0"`} {
		if !strings.Contains(got, want) {
			t.Errorf("required_providers block lacks %s:\n%s", want, got)
		}
	}
}
