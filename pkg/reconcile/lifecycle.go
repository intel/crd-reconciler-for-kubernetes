//
// Copyright (c) 2018 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: EPL-2.0
//

package reconcile

type lifecycle string

const (
	exists       lifecycle = "Exists"
	doesNotExist lifecycle = "Does-not-exist"
	deleting     lifecycle = "Deleting"
)

// isOneOf returns true if this lifecycle is in the supplied list.
func (l lifecycle) isOneOf(targets ...lifecycle) bool {
	for _, target := range targets {
		if l == target {
			return true
		}
	}
	return false
}
