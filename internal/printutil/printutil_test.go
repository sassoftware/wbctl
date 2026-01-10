// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package printutil_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	snapshot1 "github.com/kubernetes-csi/external-snapshotter/client/v8/apis/volumesnapshot/v1"
	"github.com/sassoftware/sas-ggdk/pkg/pointer"
	"github.com/sassoftware/wbctl/internal/printutil"
	"github.com/sassoftware/wbctl/internal/timeutil"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	expectedDescribePodStatusEmpty    = ""
	expectedDescribePodStatusFallback = `Pod pod-fallback:
  Phase:   Pending
  IPs:     10.0.0.9
`
	expectedDescribePodStatusPopulated = `Pod pod-a:
  Phase:            Running
  Deleted:          true
  Message:          pod message
  Reason:           pod reason
  Nominated node:   node-1
  IPs:              10.0.0.1, 10.0.0.2
  Conditions:
    Ready   True
  Init Containers:
    init-1:
      Image:           img-init
      Ready:           false
      Started:         true
      Restart Count:   0
      State:           Waiting
        Reason:   Init
  Containers:
    main-1:
      Image:           img-main
      Ready:           false
      Started:         true
      Restart Count:   1
      State:           Running
        Started At:   2024-07-15T12:00:00Z
  Ephemeral Containers:
    ephemeral-1:
      Image:           img-ephemeral
      Ready:           false
      Started:         false
      Restart Count:   0
      State:           Terminated
        Reason:        Error
        Message:       terminated
        Exit Code:     1
        Started At:    2024-07-15T12:01:00Z (9m ago)
        Finished At:   2024-07-15T12:02:00Z (8m ago)
`
	expectedWriteTabbedStorageSpecAllPopulated = `x Access Modes:    ReadWriteMany
x Storage Class:   standard
x Size:            2Gi
`

	expectedWriteTabbedStorageSpecNoStorageClass = `Access Modes:   ReadWriteOnce
Size:           1Gi
`

	expectedAppendTabbedStringOutput        = "test output"
	expectedTimePtrToStringWithAgeUnknown   = "<unknown>"
	expectedTimePtrToStringWithAgeFormatted = "2024-07-15T12:00:00Z (10m ago)"
	expectedTimePtrToStringUnknown          = "<unknown>"
	expectedTimePtrToStringFormatted        = "2024-07-15T12:00:00Z"
	expectedQuantityPtrToStringUnknown      = "<unknown>"
	expectedQuantityPtrToStringFormatted    = "1Gi"

	// AppendDataplaneDetails expected output.
	expectedAppendDataplaneDetailsBasic = `ID:            dp-123
Name:          test-dataplane
Created By:    
Created At:    <unknown>
Modified At:   <unknown>
State:         Active
Spec:          {
                 "field1": "value1",
                 "field2": {
                   "sub-field1": "sub-value1",
                   "sub-field2": "sub-value2"
                 }
               }
`
	expectedAppendDataplaneDetailsEmptySpec = `ID:            dp-123
Name:          test-dataplane
Created By:    
Created At:    <unknown>
Modified At:   <unknown>
State:         Active
`
	expectedAppendDataplaneDetailsMarshalFailure = `ID:            dp-123
Name:          test-dataplane
Created By:    
Created At:    <unknown>
Modified At:   <unknown>
State:         Active
Spec:          <error: marshal error>
`

	// AppendImageDetails expected output.
	expectedAppendImageDetailsBasic = `ID:              img-123
Controller ID:   
Name:            test-image
Created By:       ()
Created At:      <unknown>
Modified At:     <unknown>
Image:           registry.example.com/app:latest
Spec:            {
                 "field1": "value1",
                 "field2": {
                   "sub-field1": "sub-value1",
                   "sub-field2": "sub-value2"
                 }
               }
`
	expectedAppendImageDetailsEmptySpec = `ID:              img-123
Controller ID:   
Name:            test-image
Created By:       ()
Created At:      <unknown>
Modified At:     <unknown>
Image:           registry.example.com/app
`
	expectedAppendImageDetailsMarshalFailure = `ID:              img-123
Controller ID:   
Name:            test-image
Created By:       ()
Created At:      <unknown>
Modified At:     <unknown>
Image:           registry.example.com/app
Spec:            <error: marshal error>
`

	// AppendJobSubmissionDetails expected output.
	expectedAppendJobSubmissionDetailsBasic = `ID:              js-123
Controller ID:   
Job Name:        test-job
Job ID:          job-456
URL:             
Log URL:         
Created By:       ()
Created At:      <unknown>
Modified At:     <unknown>
Canceled At:     2024-07-15T12:00:00Z (10m ago)
Completed At:    2024-07-15T12:00:00Z (10m ago)
Runtime:         0.000000
Scheduled:       false
State:           
`

	// AppendUserDetails expected output.
	expectedAppendUserDetailsBasic = `ID:            auth0|user-123
Database ID:   user-123
Created At:    1970-01-01T00:00:00Z (56y ago)
Username:      testuser
Email:         test@example.com
`

	// AppendWorkspaceDetails expected output.
	expectedAppendWorkspaceDetailsBasic = `ID:              ws-123
Controller ID:   
Name:            test-workspace
Storage Type:    
Created By:       ()
Created At:      <unknown>
Modified At:     <unknown>
State:           
Phase:           
`

	// AppendWorkbenchDetails expected output.
	expectedAppendWorkbenchDetailsBasic = `ID:              wb-123
Controller ID:   
Name:            test-workbench
URL:             https://example.com
Created By:       ()
Created At:      <unknown>
Modified At:     <unknown>
Nodepool:        
Desired State:   
Current State:   
`

	// AppendLabelDetails expected output - with labels.
	expectedAppendLabelDetailsWithLabels = `Labels:
  app:   myapp
  env:   test
`
	// AppendLabelDetails expected output - empty.
	expectedAppendLabelDetailsEmpty = ""

	// AppendJobDetails expected output.
	expectedAppendJobDetailsBasic = `ID:                       job-123
Name:                     test-job
Created By:               testuser (test@example.com)
Created At:               2024-07-15T12:00:00Z (10m ago)
Modified At:              <unknown>
Most Recent Job Run At:   2024-07-15T12:00:00Z
Nodepool:                 default
Spec:                     {
                 "field1": "value1",
                 "field2": {
                   "sub-field1": "sub-value1",
                   "sub-field2": "sub-value2"
                 }
               }
`
	expectedAppendJobDetailsEmptySpec = `ID:                       job-123
Name:                     test-job
Created By:               testuser (test@example.com)
Created At:               2024-07-15T12:00:00Z (10m ago)
Modified At:              <unknown>
Most Recent Job Run At:   2024-07-15T12:00:00Z
Nodepool:                 default
`

	// AppendSnapshotDetails expected output.
	expectedAppendSnapshotDetailsBasic = `ID:              snap-123
Controller ID:   
Name:            test-snapshot
Description:     
Created By:       ()
Created At:      <unknown>
Modified At:     <unknown>
Spec:            {
                   "source": {
                     "persistentVolumeClaimName": "pvc-123",
                     "volumeSnapshotContentName": "snap-123"
                   },
                   "volumeSnapshotClassName": "test-snapshot-class"
                 }
`
	expectedAppendSnapshotDetailsEmptySpec = `ID:              snap-123
Controller ID:   
Name:            test-snapshot
Description:     
Created By:       ()
Created At:      <unknown>
Modified At:     <unknown>
`
	expectedAppendSnapshotDetailsMarshalFailure = `ID:              snap-123
Controller ID:   
Name:            test-snapshot
Description:     
Created By:       ()
Created At:      <unknown>
Modified At:     <unknown>
Spec:            <error: marshal error>
`

	// AppendEnvDetails expected output - with env.
	expectedAppendEnvDetailsWithEnv = `Environment:
  VAR1:   value1
  VAR2:   value2
`
	// AppendEnvDetails expected output - empty.
	expectedAppendEnvDetailsEmpty = ""

	// AppendPropertiesDetails expected output - with properties.
	// Order is now sorted (prop1, prop2).
	expectedAppendPropertiesDetailsWithProps = `Properties:
  prop1:   value1
  prop2:   value2
`
	// AppendPropertiesDetails expected output - empty.
	expectedAppendPropertiesDetailsEmpty = ""

	// AppendConflictDetails expected output.
	expectedAppendConflictDetailsBasic = `ID:                        conflict-1
Controller ID:             ctrl-1
Name:                      test-conflict
Type:                      resource
Conflict ID:               conflict-sub-1
Conflict Controller ID:    conflict-ctrl-1
Conflict Name:             conflicted-resource
Conflict Type:             deployment
Workspace ID:              ws-1
Workspace Controller ID:   ws-ctrl-1
Workspace Name:            test-workspace
`

	// AppendImageBuildDetails expected output.
	expectedAppendImageBuildDetailsBasic = `ID:              img-build-1
Controller ID:   ctrl-1
Name:            test-build
Created By:      user1
Created At:      2024-07-15T12:00:00Z (10m ago)
Image ID:        img-1
`

	// AppendRolesDetails expected output - with roles.
	expectedAppendRolesDetailsWithRoles = `Roles:
  - admin
  - user
`
	// AppendRolesDetails expected output - empty.
	expectedAppendRolesDetailsEmpty = ""

	// AppendServiceRequestDetails expected output - with services.
	expectedAppendServiceRequestDetailsWithServices = `Services::
  0:
    ID:      svc-1
    Image:   nginx:latest
    Environment:
      PORT:   8080
`
	// AppendServiceRequestDetails expected output - empty.
	expectedAppendServiceRequestDetailsEmpty = ""

	// AppendMetadataDetails expected output - with settings.
	// Note: order is key2, key1 due to map iteration.
	expectedAppendSettingsDetailsWithSettings = `Metadata:
  key1:   value1
  key2:   value2
`
	// AppendMetadataDetails expected output - empty.
	expectedAppendSettingsDetailsEmpty2 = ""

	// AppendVolumeDetails expected output.
	expectedAppendVolumeDetailsBasic = `Volume Status:
  Ready To Use:    true
  Restore Size:    <unknown>
  Content Name:    
  Creation Time:   <unknown>
`
	// AppendVolumeDetails expected output - nil case.
	expectedAppendVolumeDetailsNil = `Volume Status:
  Ready To Use:    false
  Restore Size:    <unknown>
  Content Name:    <unknown>
  Creation Time:   <unknown>
`

	// AppendWorkspaceRequestsDetails expected output - with workspaces.
	expectedAppendWorkspaceRequestsDetailsWithWorkspaces = `Volumes:
  0:
    ID:           ws-req-1
    Mount name:   workspace-1
`
	expectedAppendWorkspaceRequestsDetailsWithEphemeralPVC = `Volumes:
  0:
    Mount name:   workspace-1
    Ephemeral PVC:
      Access Modes:    ReadWriteOnce
      Storage Class:   standard
      Size:            1Gi
`

	// AppendWorkspaceRequestsDetails expected output - empty.
	expectedAppendWorkspaceRequestsDetailsEmpty = ""

	// AppendContainerStatuses expected outputs.
	expectedAppendContainerStatusesRunning = `  Containers
    test-container:
      Image:           test:latest
      Ready:           true
      Started:         true
      Restart Count:   1
      State:           Running
        Started At:   2024-07-15T12:00:00Z
      Limits:
        CPU:      100m
        Memory:   128Mi
      Requests:
        CPU:      50m
        Memory:   64Mi
      Allocated:
        CPU:      100m
        Memory:   128Mi
      Volume Mounts:
        /mount1 (RO)
        /mount2 
      Linux User:
        UID:               1111
        GID:               1111
        Additional GIDs:   1, 2, 3
`
	expectedAppendContainerStatusesEmpty   = ""
	expectedAppendContainerStatusesWaiting = `  Init Containers
    init-container:
      Image:           init:latest
      Ready:           false
      Started:         false
      Restart Count:   0
      State:           Waiting
        Reason:    PullBackOff
        Message:   Failed to pull image
`
	expectedAppendContainerStatusesTerminated = `  Ephemeral Containers
    ephemeral:
      Image:           ephemeral:latest
      Ready:           false
      Started:         false
      Restart Count:   0
      State:           Terminated
        Reason:        Error
        Message:       Container failed
        Exit Code:     1
        Started At:    <unknown>
        Finished At:   <unknown>
`
	expectedAppendContainerStatusesUnknown = `  Ephemeral Containers
    ephemeral:
      Image:           ephemeral:latest
      Ready:           false
      Started:         false
      Restart Count:   0
      State:           <unknown>
`
)

func TestDescribePodStatus(t *testing.T) {
	fixedNow := time.Date(2024, 7, 15, 12, 10, 0, 0, time.UTC)
	timeutil.SetNow(func() time.Time { return fixedNow })
	t.Cleanup(func() { timeutil.SetNow(time.Now) })

	tests := []struct {
		name           string
		podStatuses    []types.ExecuteableResourceStatus
		expectedOutput string
	}{
		{
			name:           "returns empty when no pod statuses",
			podStatuses:    nil,
			expectedOutput: expectedDescribePodStatusEmpty,
		},
		{
			name: "renders pod status with deleted and conditions",
			podStatuses: []types.ExecuteableResourceStatus{
				{
					Name:    "pod-a",
					Deleted: true,
					Status: corev1.PodStatus{
						Phase:             corev1.PodRunning,
						Message:           "pod message",
						Reason:            "pod reason",
						NominatedNodeName: "node-1",
						PodIP:             "10.0.0.1",
						PodIPs: []corev1.PodIP{
							{IP: "10.0.0.1"},
							{IP: "10.0.0.2"},
						},
						Conditions: []corev1.PodCondition{
							{
								Type:   corev1.PodReady,
								Status: corev1.ConditionTrue,
							},
						},
						InitContainerStatuses: []corev1.ContainerStatus{
							{
								Name:         "init-1",
								Image:        "img-init",
								Started:      pointer.Ptr(true),
								RestartCount: 0,
								State: corev1.ContainerState{
									Waiting: &corev1.ContainerStateWaiting{Reason: "Init"},
								},
							},
						},
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name:         "main-1",
								Image:        "img-main",
								Started:      pointer.Ptr(true),
								RestartCount: 1,
								State: corev1.ContainerState{
									Running: &corev1.ContainerStateRunning{
										StartedAt: metav1.NewTime(time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)),
									},
								},
							},
						},
						EphemeralContainerStatuses: []corev1.ContainerStatus{
							{
								Name:         "ephemeral-1",
								Image:        "img-ephemeral",
								Started:      pointer.Ptr(false),
								RestartCount: 0,
								State: corev1.ContainerState{
									Terminated: &corev1.ContainerStateTerminated{
										Reason:     "Error",
										Message:    "terminated",
										ExitCode:   1,
										StartedAt:  metav1.NewTime(time.Date(2024, 7, 15, 12, 1, 0, 0, time.UTC)),
										FinishedAt: metav1.NewTime(time.Date(2024, 7, 15, 12, 2, 0, 0, time.UTC)),
									},
								},
							},
						},
					},
				},
			},
			expectedOutput: expectedDescribePodStatusPopulated,
		},
		{
			name: "renders pod IP fallback when pod IP list is empty",
			podStatuses: []types.ExecuteableResourceStatus{
				{
					Name: "pod-fallback",
					Status: corev1.PodStatus{
						Phase: corev1.PodPending,
						PodIP: "10.0.0.9",
					},
				},
			},
			expectedOutput: expectedDescribePodStatusFallback,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var actual strings.Builder
			printutil.AppendPodStatusDetails(&actual, testCase.podStatuses)
			require.Equal(t, testCase.expectedOutput, actual.String())
		})
	}
}

func TestWriteTabbedStorageSpec(t *testing.T) {
	tests := []struct {
		name           string
		spec           *corev1.PersistentVolumeClaimSpec
		prefix         string
		expectedOutput string
	}{
		{
			name: "writes all populated fields",
			spec: &corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("2Gi"),
					},
				},
				StorageClassName: pointer.Ptr("standard"),
			},
			prefix:         "x ",
			expectedOutput: expectedWriteTabbedStorageSpecAllPopulated,
		},
		{
			name: "omits storage class when not set",
			spec: &corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("1Gi"),
					},
				},
				StorageClassName: nil,
			},
			prefix:         "",
			expectedOutput: expectedWriteTabbedStorageSpecNoStorageClass,
		},
		{
			name:           "nil appends nothing",
			spec:           nil,
			prefix:         "",
			expectedOutput: "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			printutil.AppendStorageSpecDetails(buffer, testCase.spec, testCase.prefix)
			require.Equal(t, testCase.expectedOutput, buffer.String())
		})
	}
}

func TestAppendTabbedString(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(io.Writer)
		expected string
	}{
		{
			name: "appends formatted output",
			fn: func(w io.Writer) {
				_, _ = w.Write([]byte("test output"))
			},
			expected: expectedAppendTabbedStringOutput,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var buf bytes.Buffer
			printutil.AppendTabbedString(&buf, testCase.fn)
			require.Equal(t, testCase.expected, buf.String())
		})
	}
}

func TestTimePtrToStringWithAge(t *testing.T) {
	tests := []struct {
		name     string
		input    *time.Time
		expected string
	}{
		{
			name:     "returns unknown when nil",
			input:    nil,
			expected: expectedTimePtrToStringWithAgeUnknown,
		},
		{
			name:     "formats time as RFC3339",
			input:    pointer.Ptr(time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)),
			expected: expectedTimePtrToStringWithAgeFormatted,
		},
	}

	for _, testCase := range tests {
		fixedNow := time.Date(2024, 7, 15, 12, 10, 0, 0, time.UTC)
		timeutil.SetNow(func() time.Time { return fixedNow })
		t.Cleanup(func() { timeutil.SetNow(time.Now) })
		t.Run(testCase.name, func(t *testing.T) {
			result := printutil.TimePtrToStringWithAge(testCase.input)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestTimePtrToString(t *testing.T) {
	tests := []struct {
		name     string
		input    *time.Time
		expected string
	}{
		{
			name:     "returns unknown when nil",
			input:    nil,
			expected: expectedTimePtrToStringUnknown,
		},
		{
			name:     "formats time as RFC3339",
			input:    pointer.Ptr(time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)),
			expected: expectedTimePtrToStringFormatted,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := printutil.TimePtrToString(testCase.input)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestQuantityPtrAsString(t *testing.T) {
	tests := []struct {
		name     string
		input    *resource.Quantity
		expected string
	}{
		{
			name:     "returns unknown when nil",
			input:    nil,
			expected: expectedQuantityPtrToStringUnknown,
		},
		{
			name:     "formats quantity as string",
			input:    pointer.Ptr(resource.MustParse("1Gi")),
			expected: expectedQuantityPtrToStringFormatted,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := printutil.QuantityPtrToString(testCase.input)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestAppendDataplaneDetails(t *testing.T) {
	fixedNow := time.Date(2024, 7, 15, 12, 10, 0, 0, time.UTC)
	timeutil.SetNow(func() time.Time { return fixedNow })
	t.Cleanup(func() { timeutil.SetNow(time.Now) })

	tests := []struct {
		name           string
		resource       types.Dataplane
		expected       string
		marshalFailure bool
	}{
		{
			name: "appends dataplane details",
			resource: types.Dataplane{
				ID:    "dp-123",
				Name:  "test-dataplane",
				State: "Active",
				Spec: map[string]interface{}{
					"field1": "value1",
					"field2": map[string]interface{}{
						"sub-field1": "sub-value1",
						"sub-field2": "sub-value2",
					},
				},
			},
			expected: expectedAppendDataplaneDetailsBasic,
		},
		{
			name: "appends dataplane details - empty spec",
			resource: types.Dataplane{
				ID:    "dp-123",
				Name:  "test-dataplane",
				State: "Active",
			},
			expected: expectedAppendDataplaneDetailsEmptySpec,
		},
		{
			name: "appends dataplane details - marshal failure",
			resource: types.Dataplane{
				ID:    "dp-123",
				Name:  "test-dataplane",
				State: "Active",
				Spec: map[string]interface{}{
					"field1": "value1",
					"field2": map[string]interface{}{
						"sub-field1": "sub-value1",
						"sub-field2": "sub-value2",
					},
				},
			},
			expected:       expectedAppendDataplaneDetailsMarshalFailure,
			marshalFailure: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.marshalFailure {
				printutil.SetMarshalIndentedFunc(func(any, string, string) ([]byte, error) {
					return nil, errors.New("marshal error")
				})
				t.Cleanup(func() { printutil.SetMarshalIndentedFunc(json.MarshalIndent) })
			}
			var sb strings.Builder
			printutil.AppendDataplaneDetails(&sb, testCase.resource)
			require.Equal(t, testCase.expected, sb.String())
		})
	}
}

func TestAppendImageDetails(t *testing.T) {
	fixedNow := time.Date(2024, 7, 15, 12, 10, 0, 0, time.UTC)
	timeutil.SetNow(func() time.Time { return fixedNow })
	t.Cleanup(func() { timeutil.SetNow(time.Now) })

	tests := []struct {
		name           string
		resource       types.Image
		expected       string
		marshalFailure bool
	}{
		{
			name: "appends image details",
			resource: types.Image{
				ID:            "img-123",
				Name:          "test-image",
				ImageTag:      "latest",
				ImageName:     "app",
				ImageRegistry: "registry.example.com",
				Spec: map[string]interface{}{
					"field1": "value1",
					"field2": map[string]interface{}{
						"sub-field1": "sub-value1",
						"sub-field2": "sub-value2",
					},
				},
			},
			expected: expectedAppendImageDetailsBasic,
		},
		{
			name: "appends image details - empty spec",
			resource: types.Image{
				ID:            "img-123",
				Name:          "test-image",
				ImageName:     "app",
				ImageRegistry: "registry.example.com",
			},
			expected: expectedAppendImageDetailsEmptySpec,
		},
		{
			name: "appends image details - marshal failure",
			resource: types.Image{
				ID:            "img-123",
				Name:          "test-image",
				ImageName:     "app",
				ImageRegistry: "registry.example.com",
				Spec: map[string]interface{}{
					"field1": "value1",
					"field2": map[string]interface{}{
						"sub-field1": "sub-value1",
						"sub-field2": "sub-value2",
					},
				},
			},
			expected:       expectedAppendImageDetailsMarshalFailure,
			marshalFailure: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.marshalFailure {
				printutil.SetMarshalIndentedFunc(func(any, string, string) ([]byte, error) {
					return nil, errors.New("marshal error")
				})
				t.Cleanup(func() { printutil.SetMarshalIndentedFunc(json.MarshalIndent) })
			}
			var sb strings.Builder
			printutil.AppendImageDetails(&sb, testCase.resource)
			require.Equal(t, testCase.expected, sb.String())
		})
	}
}

func TestAppendJobSubmissionDetails(t *testing.T) {
	fixedNow := time.Date(2024, 7, 15, 12, 10, 0, 0, time.UTC)
	timeutil.SetNow(func() time.Time { return fixedNow })
	t.Cleanup(func() { timeutil.SetNow(time.Now) })

	tests := []struct {
		name     string
		resource types.JobSubmission
		expected string
	}{
		{
			name: "appends job submission details",
			resource: types.JobSubmission{
				ID:          "js-123",
				JobName:     "test-job",
				JobID:       "job-456",
				CompletedAt: pointer.Ptr(time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)),
				CanceledAt:  pointer.Ptr(time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)),
			},
			expected: expectedAppendJobSubmissionDetailsBasic,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var sb strings.Builder
			printutil.AppendJobSubmissionDetails(&sb, testCase.resource)
			require.Equal(t, testCase.expected, sb.String())
		})
	}
}

func TestAppendUserDetails(t *testing.T) {
	tests := []struct {
		name     string
		user     types.Auth0User
		expected string
	}{
		{
			name: "appends user details",
			user: types.Auth0User{
				ID:                "auth0|user-123",
				DatabaseID:        "user-123",
				Name:              "testuser",
				Email:             "test@example.com",
				CreationTimeStamp: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expected: expectedAppendUserDetailsBasic,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var sb strings.Builder
			printutil.AppendUserDetails(&sb, testCase.user)
			require.Equal(t, testCase.expected, sb.String())
		})
	}
}

func TestAppendWorkspaceDetails(t *testing.T) {
	fixedNow := time.Date(2024, 7, 15, 12, 10, 0, 0, time.UTC)
	timeutil.SetNow(func() time.Time { return fixedNow })
	t.Cleanup(func() { timeutil.SetNow(time.Now) })

	tests := []struct {
		name     string
		resource types.Workspace
		expected string
	}{
		{
			name: "appends workspace details",
			resource: types.Workspace{
				ID:   "ws-123",
				Name: "test-workspace",
			},
			expected: expectedAppendWorkspaceDetailsBasic,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var sb strings.Builder
			printutil.AppendWorkspaceDetails(&sb, testCase.resource)
			require.Equal(t, testCase.expected, sb.String())
		})
	}
}

func TestAppendWorkbenchDetails(t *testing.T) {
	fixedNow := time.Date(2024, 7, 15, 12, 10, 0, 0, time.UTC)
	timeutil.SetNow(func() time.Time { return fixedNow })
	t.Cleanup(func() { timeutil.SetNow(time.Now) })

	tests := []struct {
		name     string
		resource types.Workbench
		expected string
	}{
		{
			name: "appends workbench details",
			resource: types.Workbench{
				ID:   "wb-123",
				Name: "test-workbench",
				URL:  "https://example.com",
			},
			expected: expectedAppendWorkbenchDetailsBasic,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var sb strings.Builder
			printutil.AppendWorkbenchDetails(&sb, testCase.resource)
			require.Equal(t, testCase.expected, sb.String())
		})
	}
}

func TestAppendLabelDetails(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		expected string
	}{
		{
			name: "appends label details when labels exist",
			labels: map[string]string{
				"env": "test",
				"app": "myapp",
			},
			expected: expectedAppendLabelDetailsWithLabels,
		},
		{
			name:     "returns empty when no labels",
			labels:   map[string]string{},
			expected: expectedAppendLabelDetailsEmpty,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var sb strings.Builder
			printutil.AppendLabelDetails(&sb, testCase.labels)
			require.Equal(t, testCase.expected, sb.String())
		})
	}
}

func TestAppendJobDetails(t *testing.T) {
	fixedNow := time.Date(2024, 7, 15, 12, 10, 0, 0, time.UTC)
	timeutil.SetNow(func() time.Time { return fixedNow })
	t.Cleanup(func() { timeutil.SetNow(time.Now) })

	tests := []struct {
		name     string
		resource types.Job
		expected string
	}{
		{
			name: "appends job details",
			resource: types.Job{
				ID:                        "job-123",
				Name:                      "test-job",
				Nodepool:                  "default",
				Username:                  "testuser",
				Email:                     "test@example.com",
				CreatedAt:                 pointer.Ptr(time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)),
				MostRecentJobSubmissionAt: pointer.Ptr(time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)),
				Spec: map[string]interface{}{
					"field1": "value1",
					"field2": map[string]interface{}{
						"sub-field1": "sub-value1",
						"sub-field2": "sub-value2",
					},
				},
			},
			expected: expectedAppendJobDetailsBasic,
		},
		{
			name: "appends job details - empty spec",
			resource: types.Job{
				ID:                        "job-123",
				Name:                      "test-job",
				Nodepool:                  "default",
				Username:                  "testuser",
				Email:                     "test@example.com",
				CreatedAt:                 pointer.Ptr(time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)),
				MostRecentJobSubmissionAt: pointer.Ptr(time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)),
			},
			expected: expectedAppendJobDetailsEmptySpec,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var sb strings.Builder
			printutil.AppendJobDetails(&sb, testCase.resource)
			require.Equal(t, testCase.expected, sb.String())
		})
	}
}

func TestAppendSnapshotDetails(t *testing.T) {
	fixedNow := time.Date(2024, 7, 15, 12, 10, 0, 0, time.UTC)
	timeutil.SetNow(func() time.Time { return fixedNow })
	t.Cleanup(func() { timeutil.SetNow(time.Now) })

	tests := []struct {
		name           string
		resource       types.Snapshot
		expected       string
		marshalFailure bool
	}{
		{
			name: "appends snapshot details",
			resource: types.Snapshot{
				ID:   "snap-123",
				Name: "test-snapshot",
				Spec: &snapshot1.VolumeSnapshotSpec{
					Source: snapshot1.VolumeSnapshotSource{
						PersistentVolumeClaimName: pointer.Ptr("pvc-123"),
						VolumeSnapshotContentName: pointer.Ptr("snap-123"),
					},
					VolumeSnapshotClassName: pointer.Ptr("test-snapshot-class"),
				},
			},
			expected: expectedAppendSnapshotDetailsBasic,
		},
		{
			name: "appends snapshot details - empty spec",
			resource: types.Snapshot{
				ID:   "snap-123",
				Name: "test-snapshot",
			},
			expected: expectedAppendSnapshotDetailsEmptySpec,
		},
		{
			name: "appends snapshot details - marshal fail",
			resource: types.Snapshot{
				ID:   "snap-123",
				Name: "test-snapshot",
				Spec: &snapshot1.VolumeSnapshotSpec{
					Source: snapshot1.VolumeSnapshotSource{
						PersistentVolumeClaimName: pointer.Ptr("pvc-123"),
						VolumeSnapshotContentName: pointer.Ptr("snap-123"),
					},
					VolumeSnapshotClassName: pointer.Ptr("test-snapshot-class"),
				},
			},
			expected:       expectedAppendSnapshotDetailsMarshalFailure,
			marshalFailure: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.marshalFailure {
				printutil.SetMarshalIndentedFunc(func(any, string, string) ([]byte, error) {
					return nil, errors.New("marshal error")
				})
				t.Cleanup(func() { printutil.SetMarshalIndentedFunc(json.MarshalIndent) })
			}
			var sb strings.Builder
			printutil.AppendSnapshotDetails(&sb, testCase.resource)
			require.Equal(t, testCase.expected, sb.String())
		})
	}
}

func TestAppendEnvDetails(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		expected string
	}{
		{
			name: "appends environment details when env exists",
			env: map[string]string{
				"VAR1": "value1",
				"VAR2": "value2",
			},
			expected: expectedAppendEnvDetailsWithEnv,
		},
		{
			name:     "returns empty when no env",
			env:      map[string]string{},
			expected: expectedAppendEnvDetailsEmpty,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var sb strings.Builder
			printutil.AppendEnvDetails(&sb, testCase.env)
			require.Equal(t, testCase.expected, sb.String())
		})
	}
}

func TestAppendPropertiesDetails(t *testing.T) {
	tests := []struct {
		name       string
		properties map[string]interface{}
		expected   string
	}{
		{
			name: "appends properties details when properties exist",
			properties: map[string]interface{}{
				"prop1": "value1",
				"prop2": "value2",
			},
			expected: expectedAppendPropertiesDetailsWithProps,
		},
		{
			name:       "returns empty when no properties",
			properties: map[string]interface{}{},
			expected:   expectedAppendPropertiesDetailsEmpty,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var sb strings.Builder
			printutil.AppendPropertiesDetails(&sb, testCase.properties)
			require.Equal(t, testCase.expected, sb.String())
		})
	}
}

func TestAppendConflictDetails(t *testing.T) {
	tests := []struct {
		name     string
		resource types.Conflict
		expected string
	}{
		{
			name: "appends all conflict detail fields",
			resource: types.Conflict{
				ID:                    "conflict-1",
				ControllerID:          "ctrl-1",
				Name:                  "test-conflict",
				Type:                  "resource",
				ConflictID:            "conflict-sub-1",
				ConflictControllerID:  "conflict-ctrl-1",
				ConflictName:          "conflicted-resource",
				ConflictType:          "deployment",
				WorkspaceID:           "ws-1",
				WorkspaceControllerID: "ws-ctrl-1",
				WorkspaceName:         "test-workspace",
			},
			expected: expectedAppendConflictDetailsBasic,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var sb strings.Builder
			printutil.AppendConflictDetails(&sb, testCase.resource)
			require.Equal(t, testCase.expected, sb.String())
		})
	}
}

func TestAppendImageBuildDetails(t *testing.T) {
	fixedNow := time.Date(2024, 7, 15, 12, 10, 0, 0, time.UTC)
	timeutil.SetNow(func() time.Time { return fixedNow })
	t.Cleanup(func() { timeutil.SetNow(time.Now) })

	tests := []struct {
		name     string
		resource types.ImageBuild
		expected string
	}{
		{
			name: "appends all image build detail fields",
			resource: types.ImageBuild{
				ID:           "img-build-1",
				ControllerID: "ctrl-1",
				Name:         "test-build",
				CreatedBy:    "user1",
				CreatedAt:    pointer.Ptr(fixedNow.Add(-10 * time.Minute)),
				ImageID:      "img-1",
			},
			expected: expectedAppendImageBuildDetailsBasic,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var sb strings.Builder
			printutil.AppendImageBuildDetails(&sb, testCase.resource)
			require.Equal(t, testCase.expected, sb.String())
		})
	}
}

func TestAppendRolesDetails(t *testing.T) {
	tests := []struct {
		name     string
		roles    []types.Role
		expected string
	}{
		{
			name: "appends roles when present",
			roles: []types.Role{
				{Name: "admin"},
				{Name: "user"},
			},
			expected: expectedAppendRolesDetailsWithRoles,
		},
		{
			name:     "returns empty when no roles",
			roles:    []types.Role{},
			expected: expectedAppendRolesDetailsEmpty,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var sb strings.Builder
			printutil.AppendRolesDetails(&sb, testCase.roles)
			require.Equal(t, testCase.expected, sb.String())
		})
	}
}

func TestAppendServiceRequestDetails(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		services []types.ServiceRequest
		expected string
	}{
		{
			name:   "appends services with header",
			header: "Services:",
			services: []types.ServiceRequest{
				{
					ID:    "svc-1",
					Image: "nginx:latest",
					Env: map[string]string{
						"PORT": "8080",
					},
				},
			},
			expected: expectedAppendServiceRequestDetailsWithServices,
		},
		{
			name:     "returns empty when no services",
			header:   "Services:",
			services: []types.ServiceRequest{},
			expected: expectedAppendServiceRequestDetailsEmpty,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var sb strings.Builder
			printutil.AppendServiceRequestDetails(&sb, testCase.header, testCase.services)
			require.Equal(t, testCase.expected, sb.String())
		})
	}
}

func TestAppendSettingsDetails(t *testing.T) {
	tests := []struct {
		name     string
		settings map[string]any
		expected string
	}{
		{
			name: "appends settings when present",
			settings: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
			expected: expectedAppendSettingsDetailsWithSettings,
		},
		{
			name:     "returns empty when no settings",
			settings: map[string]any{},
			expected: expectedAppendSettingsDetailsEmpty2,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var sb strings.Builder
			printutil.AppendMetadataDetails(&sb, testCase.settings)
			require.Equal(t, testCase.expected, sb.String())
		})
	}
}

func TestAppendVolumeDetails(t *testing.T) {
	readyToUse := true
	tests := []struct {
		name     string
		volume   *types.VolumeStatus
		expected string
	}{
		{
			name: "appends volume details when populated",
			volume: &types.VolumeStatus{
				Name: "test-volume",
				Status: snapshot1.VolumeSnapshotStatus{
					ReadyToUse: &readyToUse,
				},
			},
			expected: expectedAppendVolumeDetailsBasic,
		},
		{
			name:     "appends volume details when nil",
			volume:   nil,
			expected: expectedAppendVolumeDetailsNil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var sb strings.Builder
			printutil.AppendVolumeDetails(&sb, testCase.volume)
			require.Equal(t, testCase.expected, sb.String())
		})
	}
}

func TestAppendWorkspaceRequestsDetails(t *testing.T) {
	tests := []struct {
		name       string
		workspaces []types.WorkspaceRequest
		expected   string
	}{
		{
			name: "appends workspace requests when present",
			workspaces: []types.WorkspaceRequest{
				{
					ID:        "ws-req-1",
					MountName: "workspace-1",
				},
			},
			expected: expectedAppendWorkspaceRequestsDetailsWithWorkspaces,
		},
		{
			name: "ephemeral workspace request",
			workspaces: []types.WorkspaceRequest{
				{
					MountName: "workspace-1",
					EphemeralPVC: &corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						StorageClassName: pointer.Ptr("standard"),
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
						VolumeMode: pointer.Ptr(corev1.PersistentVolumeFilesystem),
					},
				},
			},
			expected: expectedAppendWorkspaceRequestsDetailsWithEphemeralPVC,
		},
		{
			name:       "returns empty when no workspaces",
			workspaces: []types.WorkspaceRequest{},
			expected:   expectedAppendWorkspaceRequestsDetailsEmpty,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var sb strings.Builder
			printutil.AppendWorkspaceRequestsDetails(&sb, testCase.workspaces)
			require.Equal(t, testCase.expected, sb.String())
		})
	}
}

func TestAppendContainerStatuses(t *testing.T) {
	tests := []struct {
		name              string
		title             string
		containerStatuses []corev1.ContainerStatus
		expected          string
	}{
		{
			name:              "returns empty when no statuses",
			title:             "Containers",
			containerStatuses: []corev1.ContainerStatus{},
			expected:          expectedAppendContainerStatusesEmpty,
		},
		{
			name:  "appends container status with all fields",
			title: "Containers",
			containerStatuses: []corev1.ContainerStatus{
				{
					Name:         "test-container",
					Image:        "test:latest",
					Ready:        true,
					Started:      pointer.Ptr(true),
					RestartCount: 1,
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("50m"),
							corev1.ResourceMemory: resource.MustParse("64Mi"),
						},
					},
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{
							StartedAt: metav1.NewTime(time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)),
						},
					},
					AllocatedResources: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("128Mi"),
					},
					VolumeMounts: []corev1.VolumeMountStatus{
						{
							Name:      "test-volume1",
							MountPath: "/mount1",
							ReadOnly:  true,
						},
						{
							Name:      "test-volume2",
							MountPath: "/mount2",
							ReadOnly:  false,
						},
					},
					User: &corev1.ContainerUser{
						Linux: &corev1.LinuxContainerUser{
							UID:                1111,
							GID:                1111,
							SupplementalGroups: []int64{1, 2, 3},
						},
					},
				},
			},
			expected: expectedAppendContainerStatusesRunning,
		},
		{
			name:  "appends container status with waiting state",
			title: "Init Containers",
			containerStatuses: []corev1.ContainerStatus{
				{
					Name:         "init-container",
					Image:        "init:latest",
					Ready:        false,
					Started:      pointer.Ptr(false),
					RestartCount: 0,
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason:  "PullBackOff",
							Message: "Failed to pull image",
						},
					},
				},
			},
			expected: expectedAppendContainerStatusesWaiting,
		},
		{
			name:  "appends container status with terminated state",
			title: "Ephemeral Containers",
			containerStatuses: []corev1.ContainerStatus{
				{
					Name:         "ephemeral",
					Image:        "ephemeral:latest",
					Ready:        false,
					Started:      pointer.Ptr(false),
					RestartCount: 0,
					State: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							ExitCode: 1,
							Reason:   "Error",
							Message:  "Container failed",
						},
					},
				},
			},
			expected: expectedAppendContainerStatusesTerminated,
		},
		{
			name:  "unknown state",
			title: "Ephemeral Containers",
			containerStatuses: []corev1.ContainerStatus{
				{
					Name:         "ephemeral",
					Image:        "ephemeral:latest",
					Ready:        false,
					Started:      pointer.Ptr(false),
					RestartCount: 0,
					State:        corev1.ContainerState{},
				},
			},
			expected: expectedAppendContainerStatusesUnknown,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var sb strings.Builder
			printutil.AppendContainerStatuses(&sb, testCase.title, testCase.containerStatuses)
			require.Equal(t, testCase.expected, sb.String())
		})
	}
}
