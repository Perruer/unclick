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

import "testing"

func TestLinkByID(t *testing.T) {
	vpc := NewSimpleResource("vpc-1", "vpc-1", "aws_vpc", "aws", nil)
	vpc.Item = map[string]interface{}{"cidr_block": "10.0.0.0/16"}
	sg := NewSimpleResource("sg-1", "sg-1", "aws_security_group", "aws", nil)
	sg.Item = map[string]interface{}{"vpc_id": "vpc-1", "name": "vpc-1"}
	subnet := NewSimpleResource("subnet-1", "subnet-1", "aws_subnet", "aws", nil)
	subnet.Item = map[string]interface{}{
		"vpc_id":   "vpc-1",
		"owner_id": "123456789012",
		"rule":     []interface{}{map[string]interface{}{"security_group_ids": []interface{}{"sg-1", "sg-other"}}},
	}
	// Two resources sharing an ID make references to it ambiguous.
	a := NewSimpleResource("dup", "a", "x_thing", "x", nil)
	a.Item = map[string]interface{}{}
	b := NewSimpleResource("dup", "b", "x_other", "x", nil)
	b.Item = map[string]interface{}{"parent_id": "dup"}

	LinkByID([]*Resource{&vpc, &sg, &subnet, &a, &b})

	if got := subnet.Item["vpc_id"]; got != "${aws_vpc.tfer--vpc-1.id}" {
		t.Errorf("subnet vpc_id = %v", got)
	}
	if got := sg.Item["name"]; got != "vpc-1" {
		t.Errorf("only *_id arguments may be linked, name = %v", got)
	}
	if got := subnet.Item["owner_id"]; got != "123456789012" {
		t.Errorf("an ID of no generated resource must stay, owner_id = %v", got)
	}
	ids := subnet.Item["rule"].([]interface{})[0].(map[string]interface{})["security_group_ids"].([]interface{})
	if ids[0] != "${aws_security_group.tfer--sg-1.id}" || ids[1] != "sg-other" {
		t.Errorf("security_group_ids = %v", ids)
	}
	if got := b.Item["parent_id"]; got != "dup" {
		t.Errorf("ambiguous ID must stay, parent_id = %v", got)
	}
}
