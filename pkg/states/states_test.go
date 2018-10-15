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

package states

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsTerminal(t *testing.T) {
	testCases := []struct {
		st       State
		expected bool
	}{
		{
			st:       Completed,
			expected: true,
		},
		{
			st:       Failed,
			expected: true,
		},
		{
			st:       Pending,
			expected: false,
		},
		{
			st:       Running,
			expected: false,
		},
	}

	for _, testCase := range testCases {
		actual := IsTerminal(testCase.st)
		require.Equal(t, actual, testCase.expected)
	}
}

func TestIsOneOf(t *testing.T) {
	testCases := []struct {
		currentState State
		targetStates []State
		expected     bool
	}{
		{
			currentState: Pending,
			targetStates: []State{Pending},
			expected:     true,
		},
		{
			currentState: Pending,
			targetStates: []State{Pending, Failed},
			expected:     true,
		},
		{
			currentState: Pending,
			targetStates: []State{Failed, Pending},
			expected:     true,
		},
		{
			currentState: Pending,
			targetStates: []State{Running},
			expected:     false,
		},
		{
			currentState: Pending,
			targetStates: []State{Failed, Running},
			expected:     false,
		},
	}

	for _, testCase := range testCases {
		actual := testCase.currentState.IsOneOf(testCase.targetStates...)
		require.Equal(t, actual, testCase.expected)
	}
}
