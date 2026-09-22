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

// Package legacy holds the few Terraform 0.12 state types that the provider
// importers inherited from Terraformer still use to pass resources around.
// They are plain data: nothing here talks to Terraform.
package legacy

import "strings"

// InstanceInfo identifies a resource by its type and HCL address.
type InstanceInfo struct {
	Id   string //nolint:revive // name kept for compatibility with importers
	Type string
}

// ResourceAddress is the parsed form of InstanceInfo.Id.
type ResourceAddress struct {
	Type string
	Name string
}

// String returns the address as written in configuration, "type.name".
func (a *ResourceAddress) String() string {
	return a.Type + "." + a.Name
}

// ResourceAddress splits the HCL address of the resource.
func (i *InstanceInfo) ResourceAddress() *ResourceAddress {
	return &ResourceAddress{Type: i.Type, Name: strings.TrimPrefix(i.Id, i.Type+".")}
}

// InstanceState is a resource in the flatmap form: every value, however
// deeply nested, is a string under a dotted key such as "tags.Name".
type InstanceState struct {
	ID         string
	Attributes map[string]string
	Meta       map[string]interface{}
}

// OutputState is an output value of a generated module.
type OutputState struct {
	Sensitive bool
	Type      string
	Value     interface{}
}
