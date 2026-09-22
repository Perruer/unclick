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

package legacy

import (
	"hash/crc32"
	"os"
	"strings"
)

// HashcodeString hashes a string to a non-negative int the way Terraform
// 0.12's helper/hashcode did. Some resource IDs are derived from it, so it
// must not change.
func HashcodeString(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}

// PathOrContents returns the contents of the file at poc, or poc itself when
// it is not a path. A leading "~/" means the home directory. The boolean
// reports whether a file was read.
func PathOrContents(poc string) (string, bool, error) {
	if poc == "" {
		return poc, false, nil
	}
	path := poc
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path, true, err
		}
		path = home + path[1:]
	}
	if _, err := os.Stat(path); err == nil {
		contents, err := os.ReadFile(path)
		if err != nil {
			return string(contents), true, err
		}
		return string(contents), true, nil
	}
	return poc, false, nil
}
