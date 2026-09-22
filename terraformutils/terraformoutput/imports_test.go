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
	"testing"

	"github.com/Perruer/unclick/terraformutils"
)

func TestImportBlocks(t *testing.T) {
	sg := terraformutils.NewSimpleResource("sg-1", "web", "aws_security_group", "aws", nil)
	vpc := terraformutils.NewSimpleResource("vpc-1", "main", "aws_vpc", "aws", nil)
	odd := terraformutils.NewSimpleResource(`we"ird ${x}`, "odd", "aws_vpc", "aws", nil)
	attach := terraformutils.NewSimpleResource("role-20260101", "attach", "aws_iam_role_policy_attachment", "aws", nil)
	attach.ImportID = "role/arn:aws:iam::aws:policy/ReadOnly"

	got := string(ImportBlocks([]terraformutils.Resource{vpc, odd, sg, attach}))
	want := `import {
  to = aws_iam_role_policy_attachment.tfer--attach
  id = "role/arn:aws:iam::aws:policy/ReadOnly"
}

import {
  to = aws_security_group.tfer--web
  id = "sg-1"
}

import {
  to = aws_vpc.tfer--main
  id = "vpc-1"
}

import {
  to = aws_vpc.tfer--odd
  id = "we\"ird $${x}"
}
`
	if got != want {
		t.Fatalf("import blocks:\n%s\nwant:\n%s", got, want)
	}
}
