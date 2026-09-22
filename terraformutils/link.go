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

package terraformutils

import "strings"

// LinkByID replaces literal IDs of other generated resources with
// references, so that the configuration keeps the relationships between
// objects: `vpc_id = aws_vpc.tfer--vpc-1.id` instead of a hard-coded ID.
//
// Only arguments named "*_id" or "*_ids" are considered, and only exact
// matches of another resource's ID; an ID shared by two resources is left
// alone because the reference would be ambiguous.
func LinkByID(resources []*Resource) {
	byID := map[string]*Resource{}
	ambiguous := map[string]bool{}
	for _, r := range resources {
		id := r.InstanceState.ID
		if id == "" {
			continue
		}
		if _, seen := byID[id]; seen {
			ambiguous[id] = true
		}
		byID[id] = r
	}
	for _, r := range resources {
		link := func(value string) (string, bool) {
			target, ok := byID[value]
			if !ok || ambiguous[value] || target == r {
				return value, false
			}
			return "${" + target.InstanceInfo.Type + "." + target.ResourceName + ".id}", true
		}
		linkMap(r.Item, link)
	}
}

func linkMap(m map[string]interface{}, link func(string) (string, bool)) {
	for k, v := range m {
		isIDKey := strings.HasSuffix(k, "_id") || strings.HasSuffix(k, "_ids")
		switch t := v.(type) {
		case string:
			if isIDKey {
				if ref, ok := link(t); ok {
					m[k] = ref
				}
			}
		case []interface{}:
			for i, e := range t {
				switch et := e.(type) {
				case string:
					if isIDKey {
						if ref, ok := link(et); ok {
							t[i] = ref
						}
					}
				case map[string]interface{}:
					linkMap(et, link)
				}
			}
		case map[string]interface{}:
			linkMap(t, link)
		}
	}
}
