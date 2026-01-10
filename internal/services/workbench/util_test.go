// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench_test

import (
	"net/url"
	"testing"

	"github.com/sassoftware/wbctl/internal/services/workbench"
	"github.com/stretchr/testify/require"
)

func TestIDsOrControllerIDsQuery(t *testing.T) {
	tests := []struct {
		name     string
		ids      []string
		sort     string
		expected url.Values
	}{
		{
			name:     "empty ids",
			ids:      nil,
			expected: url.Values{},
		},
		{
			name: "all uuids",
			ids:  []string{"123e4567-e89b-12d3-a456-426614174000", "123e4567-e89b-12d3-a456-426614174001"},
			sort: "id.asc",
			expected: url.Values{
				"id":    []string{"in.(123e4567-e89b-12d3-a456-426614174000,123e4567-e89b-12d3-a456-426614174001)"},
				"order": []string{"id.asc"},
			},
		},
		{
			name:     "all controller ids",
			ids:      []string{"controller-a", "controller-b"},
			expected: url.Values{"controller_id": []string{"in.(controller-a,controller-b)"}},
		},
		{
			name: "mixed uuid and controller ids",
			ids:  []string{"123e4567-e89b-12d3-a456-426614174000", "controller-a"},
			expected: url.Values{
				"or": []string{"(id.in.(123e4567-e89b-12d3-a456-426614174000),controller_id.in.(controller-a))"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := workbench.IDsOrControllerIDsQuery(tt.ids, tt.sort)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestIDsQuery(t *testing.T) {
	tests := []struct {
		name     string
		ids      []string
		expected url.Values
		sort     string
	}{
		{
			name:     "empty ids",
			ids:      nil,
			expected: url.Values{},
		},
		{
			name:     "single id",
			ids:      []string{"id-a"},
			expected: url.Values{"id": []string{"in.(id-a)"}},
		},
		{
			name:     "multiple ids",
			ids:      []string{"id-a", "id-b"},
			expected: url.Values{"id": []string{"in.(id-a,id-b)"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := workbench.IDsQuery(tt.ids, tt.sort)
			require.Equal(t, tt.expected, actual)
		})
	}
}
