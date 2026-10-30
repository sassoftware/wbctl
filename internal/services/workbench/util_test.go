// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench_test

import (
	"net/url"
	"testing"

	"github.com/sassoftware/wbctl/internal/services/workbench"
	"github.com/stretchr/testify/require"
)

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

func TestIDsControllerIDsOrNamesQuery(t *testing.T) {
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
				"or": []string{
					"(id.in.(123e4567-e89b-12d3-a456-426614174000,123e4567-e89b-12d3-a456-426614174001),name.in.(123e4567-e89b-12d3-a456-426614174000,123e4567-e89b-12d3-a456-426614174001))",
				},
				"order": []string{"id.asc"},
			},
		},
		{
			name: "all controller ids / names",
			ids:  []string{"controller-a", "name-a"},
			expected: url.Values{
				"or": []string{
					"(controller_id.in.(controller-a,name-a),name.in.(controller-a,name-a))",
				},
			},
		},
		{
			name: "mixed uuid, controller id, names",
			ids:  []string{"123e4567-e89b-12d3-a456-426614174000", "controller-a", "name-a"},
			expected: url.Values{
				"or": []string{
					"(id.in.(123e4567-e89b-12d3-a456-426614174000),controller_id.in.(controller-a,name-a),name.in.(123e4567-e89b-12d3-a456-426614174000,controller-a,name-a))",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := workbench.IDsControllerIDsOrNamesQuery(tt.ids, tt.sort)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestIDsOrNamesQuery(t *testing.T) {
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
				"or": []string{
					"(id.in.(123e4567-e89b-12d3-a456-426614174000,123e4567-e89b-12d3-a456-426614174001),name.in.(123e4567-e89b-12d3-a456-426614174000,123e4567-e89b-12d3-a456-426614174001))",
				},
				"order": []string{"id.asc"},
			},
		},
		{
			name: "all names",
			ids:  []string{"name-a", "name-b"},
			expected: url.Values{
				"or": []string{"(name.in.(name-a,name-b))"},
			},
		},
		{
			name: "mixed uuid and names",
			ids:  []string{"123e4567-e89b-12d3-a456-426614174000", "name-a"},
			expected: url.Values{
				"or": []string{
					"(id.in.(123e4567-e89b-12d3-a456-426614174000),name.in.(123e4567-e89b-12d3-a456-426614174000,name-a))",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := workbench.IDsOrNamesQuery(tt.ids, tt.sort)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestIDsOrControllerIDsOrJobNamesQuery(t *testing.T) {
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
				"or": []string{
					"(id.in.(123e4567-e89b-12d3-a456-426614174000,123e4567-e89b-12d3-a456-426614174001),job_name.in.(123e4567-e89b-12d3-a456-426614174000,123e4567-e89b-12d3-a456-426614174001))",
				},
				"order": []string{"id.asc"},
			},
		},
		{
			name: "all controller ids / job names",
			ids:  []string{"controller-a", "job-a"},
			expected: url.Values{
				"or": []string{
					"(controller_id.in.(controller-a,job-a),job_name.in.(controller-a,job-a))",
				},
			},
		},
		{
			name: "mixed uuid, controller id, job names",
			ids:  []string{"123e4567-e89b-12d3-a456-426614174000", "controller-a", "job-a"},
			expected: url.Values{
				"or": []string{
					"(id.in.(123e4567-e89b-12d3-a456-426614174000),controller_id.in.(controller-a,job-a),job_name.in.(123e4567-e89b-12d3-a456-426614174000,controller-a,job-a))",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := workbench.IDsOrControllerIDsOrJobNamesQuery(tt.ids, tt.sort)
			require.Equal(t, tt.expected, actual)
		})
	}
}
