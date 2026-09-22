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

// Package install finds provider plugins that OpenTofu or Terraform already
// downloaded, and asks them to download missing ones.
package install

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	version "github.com/hashicorp/go-version"
)

// Registries are searched in this order. OpenTofu comes first because its
// registry mirrors almost every provider and needs no account.
var Registries = []string{"registry.opentofu.org", "registry.terraform.io"}

// Provider is a provider binary on disk.
type Provider struct {
	Source  string // "hashicorp/aws"
	Version string // "6.14.1"
	Path    string
}

// Source is the registry address of a provider. Host is empty for the
// default registry of whichever tool installs the provider.
type Source struct {
	Host      string
	Namespace string
	Type      string
}

func (s Source) String() string {
	if s.Host == "" {
		return s.Namespace + "/" + s.Type
	}
	return s.Host + "/" + s.Namespace + "/" + s.Type
}

// ParseSource accepts "type", "namespace/type" or "host/namespace/type".
// A bare type means the hashicorp namespace, as in required_providers.
func ParseSource(s string) (Source, error) {
	parts := strings.Split(strings.TrimSpace(s), "/")
	switch len(parts) {
	case 1:
		return Source{Namespace: "hashicorp", Type: parts[0]}, nil
	case 2:
		return Source{Namespace: parts[0], Type: parts[1]}, nil
	case 3:
		return Source{Host: parts[0], Namespace: parts[1], Type: parts[2]}, nil
	default:
		return Source{}, fmt.Errorf("invalid provider source %q", s)
	}
}

// SearchDirs lists the directories that may hold unpacked providers, most
// specific first: the current working directory's .terraform, the plugin
// cache, then the per-user plugin directories.
func SearchDirs() []string {
	var dirs []string
	add := func(d string) {
		if d != "" {
			dirs = append(dirs, d)
		}
	}
	dataDir := os.Getenv("TF_DATA_DIR")
	if dataDir == "" {
		dataDir = ".terraform"
	}
	add(filepath.Join(dataDir, "providers"))
	add(os.Getenv("TF_PLUGIN_CACHE_DIR"))
	if home, err := os.UserHomeDir(); err == nil {
		add(filepath.Join(home, ".terraform.d", "plugins"))
		add(filepath.Join(home, ".terraform.d", "plugin-cache"))
	}
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			add(filepath.Join(appData, "terraform.d", "plugins"))
		}
	}
	return dirs
}

// Find returns the newest installed build of the provider for this platform.
func Find(src Source, dirs []string) (*Provider, error) {
	platform := runtime.GOOS + "_" + runtime.GOARCH
	hosts := Registries
	if src.Host != "" {
		hosts = []string{src.Host}
	}
	var found []*Provider
	for _, dir := range dirs {
		for _, host := range hosts {
			base := filepath.Join(dir, host, src.Namespace, src.Type)
			versions, err := os.ReadDir(base)
			if err != nil {
				continue
			}
			for _, v := range versions {
				if !v.IsDir() {
					continue
				}
				if bin := binaryIn(filepath.Join(base, v.Name(), platform), src.Type); bin != "" {
					found = append(found, &Provider{Source: src.String(), Version: v.Name(), Path: bin})
				}
			}
		}
	}
	if len(found) == 0 {
		return nil, fmt.Errorf("provider %s is not installed", src)
	}
	sort.SliceStable(found, func(i, j int) bool { return newer(found[i].Version, found[j].Version) })
	return found[0], nil
}

func binaryIn(dir, providerType string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	prefix := "terraform-provider-" + providerType
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasPrefix(name, prefix) {
			continue
		}
		// Skip look-alikes such as terraform-provider-google-beta when
		// looking for terraform-provider-google.
		rest := strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".exe")
		if rest == "" || strings.HasPrefix(rest, "_") {
			return filepath.Join(dir, name)
		}
	}
	return ""
}

func newer(a, b string) bool {
	va, errA := version.NewVersion(a)
	vb, errB := version.NewVersion(b)
	if errA != nil || errB != nil {
		return a > b
	}
	return va.GreaterThan(vb)
}

// Installer downloads providers with the tofu or terraform binary, so that
// checksums, mirrors and credentials work exactly as the user configured.
type Installer struct {
	// Binary is the tofu or terraform executable. Empty means look for tofu,
	// then terraform, in PATH.
	Binary string
	// CacheDir holds one working directory per provider.
	CacheDir string
}

// ErrNoBinary means neither tofu nor terraform could be found.
var ErrNoBinary = errors.New("neither tofu nor terraform was found in PATH; install OpenTofu (https://opentofu.org) or pass --tf-binary")

// FindBinary returns the tofu or terraform executable to use.
func (in *Installer) FindBinary() (string, error) {
	if in.Binary != "" {
		return in.Binary, nil
	}
	for _, name := range []string{"tofu", "terraform"} {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}
	return "", ErrNoBinary
}

// Ensure returns an installed provider, running `init` in a private working
// directory when it is not installed yet. constraint may be empty.
func (in *Installer) Ensure(ctx context.Context, src Source, constraint string) (*Provider, error) {
	if constraint == "" {
		if p, err := Find(src, SearchDirs()); err == nil {
			return p, nil
		}
	}
	work := filepath.Join(in.CacheDir, src.Namespace+"-"+src.Type)
	if p, err := Find(src, []string{filepath.Join(work, ".terraform", "providers")}); err == nil && constraint == "" {
		return p, nil
	}
	bin, err := in.FindBinary()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(work, 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(work, "main.tf"), []byte(requiredProviders(src, constraint)), 0o644); err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, bin, "init", "-input=false", "-no-color", "-backend=false", "-upgrade")
	cmd.Dir = work
	cmd.Env = append(os.Environ(), "TF_IN_AUTOMATION=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("%s init for %s failed: %w\n%s", filepath.Base(bin), src, err, out)
	}
	return Find(src, []string{filepath.Join(work, ".terraform", "providers")})
}

func requiredProviders(src Source, constraint string) string {
	var b strings.Builder
	b.WriteString("terraform {\n  required_providers {\n")
	fmt.Fprintf(&b, "    %s = {\n      source = %q\n", src.Type, src.String())
	if constraint != "" {
		fmt.Fprintf(&b, "      version = %q\n", constraint)
	}
	b.WriteString("    }\n  }\n}\n")
	return b.String()
}
