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

package terraformoutput

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"github.com/Perruer/unclick/terraformutils"
)

// ImportFileName is where the import blocks for a directory are written.
const ImportFileName = "imports.tf"

// ImportBlocks renders one `import` block per resource, so that
// `tofu plan` / `terraform plan` adopts the existing objects instead of
// creating new ones. Resources are sorted by address to keep the file
// stable between runs.
func ImportBlocks(resources []terraformutils.Resource) []byte {
	sorted := make([]terraformutils.Resource, len(resources))
	copy(sorted, resources)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].InstanceInfo.Type != sorted[j].InstanceInfo.Type {
			return sorted[i].InstanceInfo.Type < sorted[j].InstanceInfo.Type
		}
		return sorted[i].ResourceName < sorted[j].ResourceName
	})

	f := hclwrite.NewEmptyFile()
	body := f.Body()
	for i, r := range sorted {
		if i > 0 {
			body.AppendNewline()
		}
		block := body.AppendNewBlock("import", nil).Body()
		block.SetAttributeTraversal("to", hcl.Traversal{
			hcl.TraverseRoot{Name: r.InstanceInfo.Type},
			hcl.TraverseAttr{Name: r.ResourceName},
		})
		block.SetAttributeValue("id", cty.StringVal(r.GetImportID()))
	}
	return f.Bytes()
}

// OutputImportFile writes the import blocks for resources into dir.
func OutputImportFile(resources []terraformutils.Resource, dir string) error {
	if len(resources) == 0 {
		return nil
	}
	return os.WriteFile(filepath.Join(dir, ImportFileName), ImportBlocks(resources), 0o644)
}
