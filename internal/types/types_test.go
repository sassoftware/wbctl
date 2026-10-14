// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"testing"

	"github.com/sassoftware/wbctl/internal/types"
	"github.com/stretchr/testify/require"
)

const (
	expectedApplicationHandler = "application-handler"
	expectedJobHandler         = "job-handler"
	resourceTypeApplications   = "applications"
	resourceTypeApplication    = "application"
	resourceTypeJob            = "job"
	resourceTypeJobs           = "jobs"
	resourceTypeMissing        = "missing"
)

func TestAsApplicationSlice(t *testing.T) {
	cases := []struct {
		name     string
		source   types.Applications
		expected []types.Application
	}{
		{
			name:     "nil slice",
			source:   nil,
			expected: nil,
		},
		{
			name: "non-empty slice",
			source: types.Applications{
				{ID: "application-1"},
			},
			expected: []types.Application{
				{ID: "application-1"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := types.AsApplicationSlice(tc.source)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestAsConflictSlice(t *testing.T) {
	cases := []struct {
		name     string
		source   types.Conflicts
		expected []types.Conflict
	}{
		{
			name:     "nil slice",
			source:   nil,
			expected: nil,
		},
		{
			name: "non-empty slice",
			source: types.Conflicts{
				{ID: "conflict-1"},
			},
			expected: []types.Conflict{
				{ID: "conflict-1"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := types.AsConflictSlice(tc.source)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestAsEventSlice(t *testing.T) {
	cases := []struct {
		name     string
		source   types.Events
		expected []types.Event
	}{
		{
			name:     "nil slice",
			source:   nil,
			expected: nil,
		},
		{
			name: "non-empty slice",
			source: types.Events{
				{ID: "event-1"},
			},
			expected: []types.Event{
				{ID: "event-1"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := types.AsEventSlice(tc.source)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestAsImageBuildLogSlice(t *testing.T) {
	cases := []struct {
		name     string
		source   types.ImageBuildLogs
		expected []types.ImageBuildLog
	}{
		{
			name:     "nil slice",
			source:   nil,
			expected: nil,
		},
		{
			name: "non-empty slice",
			source: types.ImageBuildLogs{
				{BuildID: "build-1"},
			},
			expected: []types.ImageBuildLog{
				{BuildID: "build-1"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := types.AsImageBuildLogSlice(tc.source)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestAsImageBuildSlice(t *testing.T) {
	cases := []struct {
		name     string
		source   types.ImageBuilds
		expected []types.ImageBuild
	}{
		{
			name:     "nil slice",
			source:   nil,
			expected: nil,
		},
		{
			name: "non-empty slice",
			source: types.ImageBuilds{
				{ID: "image-build-1"},
			},
			expected: []types.ImageBuild{
				{ID: "image-build-1"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := types.AsImageBuildSlice(tc.source)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestAsImageSlice(t *testing.T) {
	cases := []struct {
		name     string
		source   types.Images
		expected []types.Image
	}{
		{
			name:     "nil slice",
			source:   nil,
			expected: nil,
		},
		{
			name: "non-empty slice",
			source: types.Images{
				{ID: "image-1"},
			},
			expected: []types.Image{
				{ID: "image-1"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := types.AsImageSlice(tc.source)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestAsInitializerSlice(t *testing.T) {
	cases := []struct {
		name     string
		source   types.Initializers
		expected []types.Initializer
	}{
		{
			name:     "nil slice",
			source:   nil,
			expected: nil,
		},
		{
			name: "non-empty slice",
			source: types.Initializers{
				{ID: "initializer-1"},
			},
			expected: []types.Initializer{
				{ID: "initializer-1"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := types.AsInitializerSlice(tc.source)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestAsJobSlice(t *testing.T) {
	cases := []struct {
		name     string
		source   types.Jobs
		expected []types.Job
	}{
		{
			name:     "nil slice",
			source:   nil,
			expected: nil,
		},
		{
			name: "non-empty slice",
			source: types.Jobs{
				{ID: "job-1"},
			},
			expected: []types.Job{
				{ID: "job-1"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := types.AsJobSlice(tc.source)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestAsJobSubmissionSlice(t *testing.T) {
	cases := []struct {
		name     string
		source   types.JobSubmissions
		expected []types.JobSubmission
	}{
		{
			name:     "nil slice",
			source:   nil,
			expected: nil,
		},
		{
			name: "non-empty slice",
			source: types.JobSubmissions{
				{ID: "job-submission-1"},
			},
			expected: []types.JobSubmission{
				{ID: "job-submission-1"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := types.AsJobSubmissionSlice(tc.source)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestAsNodepoolSlice(t *testing.T) {
	cases := []struct {
		name     string
		source   types.JobSubmissions
		expected []types.JobSubmission
	}{
		{
			name:     "nil slice",
			source:   nil,
			expected: nil,
		},
		{
			name: "non-empty slice",
			source: types.JobSubmissions{
				{ID: "job-submission-1"},
			},
			expected: []types.JobSubmission{
				{ID: "job-submission-1"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := types.AsNodepoolSlice(tc.source)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestAsPingSlice(t *testing.T) {
	cases := []struct {
		name     string
		source   types.Pings
		expected []types.Ping
	}{
		{
			name:     "nil slice",
			source:   nil,
			expected: nil,
		},
		{
			name: "non-empty slice",
			source: types.Pings{
				{DataplaneID: "dataplane-1"},
			},
			expected: []types.Ping{
				{DataplaneID: "dataplane-1"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := types.AsPingSlice(tc.source)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestAsServiceSlice(t *testing.T) {
	cases := []struct {
		name     string
		source   types.Services
		expected []types.Service
	}{
		{
			name:     "nil slice",
			source:   nil,
			expected: nil,
		},
		{
			name: "non-empty slice",
			source: types.Services{
				{ID: "service-1"},
			},
			expected: []types.Service{
				{ID: "service-1"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := types.AsServiceSlice(tc.source)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestAsSnapshotSlice(t *testing.T) {
	cases := []struct {
		name     string
		source   types.Snapshots
		expected []types.Snapshot
	}{
		{
			name:     "nil slice",
			source:   nil,
			expected: nil,
		},
		{
			name: "non-empty slice",
			source: types.Snapshots{
				{ID: "snapshot-1"},
			},
			expected: []types.Snapshot{
				{ID: "snapshot-1"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := types.AsSnapshotSlice(tc.source)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestAsWorkbenchSlice(t *testing.T) {
	cases := []struct {
		name     string
		source   types.Workbenches
		expected []types.Workbench
	}{
		{
			name:     "nil slice",
			source:   nil,
			expected: nil,
		},
		{
			name: "non-empty slice",
			source: types.Workbenches{
				{ID: "workbench-1"},
			},
			expected: []types.Workbench{
				{ID: "workbench-1"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := types.AsWorkbenchSlice(tc.source)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestAsWorkspaceSlice(t *testing.T) {
	cases := []struct {
		name     string
		source   types.Workspaces
		expected []types.Workspace
	}{
		{
			name:     "nil slice",
			source:   nil,
			expected: nil,
		},
		{
			name: "non-empty slice",
			source: types.Workspaces{
				{ID: "workspace-1"},
			},
			expected: []types.Workspace{
				{ID: "workspace-1"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := types.AsWorkspaceSlice(tc.source)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestGetFunction(t *testing.T) {
	metadata := []types.FunctionMetadata[string]{
		{
			Singular: resourceTypeJob,
			Plural:   resourceTypeJobs,
			F:        expectedJobHandler,
		},
		{
			Singular: resourceTypeApplication,
			Plural:   resourceTypeApplications,
			F:        expectedApplicationHandler,
		},
	}

	tests := []struct {
		name             string
		resourceType     string
		expectedFunction string
		expectedFound    bool
	}{
		{
			name:             "finds by singular resource type",
			resourceType:     resourceTypeJob,
			expectedFunction: expectedJobHandler,
			expectedFound:    true,
		},
		{
			name:             "finds by plural resource type",
			resourceType:     resourceTypeApplications,
			expectedFunction: expectedApplicationHandler,
			expectedFound:    true,
		},
		{
			name:             "returns zero value when not found",
			resourceType:     resourceTypeMissing,
			expectedFunction: "",
			expectedFound:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualFn, actualFound := types.GetFunction(tt.resourceType, metadata)
			require.Equal(t, tt.expectedFound, actualFound)
			require.Equal(t, tt.expectedFunction, actualFn)
		})
	}
}

func TestSupportedResourceTypes(t *testing.T) {
	pluralMetadata := []types.FunctionMetadata[string]{
		{
			Singular: resourceTypeJob,
			Plural:   resourceTypeJobs,
			F:        expectedJobHandler,
		},
		{
			Singular: resourceTypeApplication,
			Plural:   resourceTypeApplications,
			F:        expectedApplicationHandler,
		},
	}

	singularMetadata := []types.FunctionMetadata[string]{
		{
			Singular: resourceTypeJob,
			F:        expectedJobHandler,
		},
		{
			Singular: resourceTypeApplication,
			F:        expectedApplicationHandler,
		},
	}

	tests := []struct {
		name     string
		metadata []types.FunctionMetadata[string]
		expected []string
	}{
		{
			name:     "returns plural names in sorted order",
			metadata: pluralMetadata,
			expected: []string{resourceTypeApplications, resourceTypeJobs},
		},
		{
			name:     "returns singular names in sorted order",
			metadata: singularMetadata,
			expected: []string{resourceTypeApplication, resourceTypeJob},
		},
		{
			name:     "returns empty slice when metadata is empty",
			metadata: nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := types.SupportedResourceTypes(tt.metadata)
			require.Equal(t, tt.expected, actual)
		})
	}
}
