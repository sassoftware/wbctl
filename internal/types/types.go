// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"slices"
	"strings"
	"time"

	snapshot1 "github.com/kubernetes-csi/external-snapshotter/client/v8/apis/volumesnapshot/v1"
	"github.com/sassoftware/wbctl/internal/maputil"
	corev1 "k8s.io/api/core/v1"
)

// Config holds CLI configuration values.
type Config struct {
	Endpoint string `json:"endpoint,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Output   string `json:"output,omitempty"`
}

// ExecuteableResourceStatus describes a resource's execution status.
type ExecuteableResourceStatus struct {
	Name    string           `json:"name,omitempty"`
	Deleted bool             `json:"deleted,omitempty"`
	NodeID  string           `json:"nodeId,omitempty"`
	Status  corev1.PodStatus `json:"status,omitempty"`
}

// WorkspaceRequest describes a workspace request payload.
type WorkspaceRequest struct {
	ID           string                            `json:"id,omitempty"`
	MountName    string                            `json:"mountName,omitempty"`
	EphemeralPVC *corev1.PersistentVolumeClaimSpec `json:"ephemeralPVC,omitempty"`
}

// ServiceRequest describes a service request payload.
type ServiceRequest struct {
	ID    string            `json:"id,omitempty"`
	Env   map[string]string `json:"env,omitempty"`
	Image string            `json:"image,omitempty"`
}

// Job represents a scheduled or ad hoc job.
type Job struct {
	ApplicationID             string                 `json:"application_id,omitempty" wbctl:"editable"`
	Context                   string                 `json:"context,omitempty" wbctl:"editable"`
	CreatedAt                 *time.Time             `json:"created_at,omitempty"`
	CreatedBy                 string                 `json:"created_by,omitempty"`
	DataplaneID               string                 `json:"dataplane_id,omitempty" wbctl:"editable"`
	Description               string                 `json:"description,omitempty" wbctl:"editable"`
	Email                     string                 `json:"email,omitempty"`
	ID                        string                 `json:"id,omitempty"`
	Labels                    map[string]string      `json:"labels,omitempty" wbctl:"editable"`
	ModifiedAt                *time.Time             `json:"modified_at,omitempty"`
	ModifiedBy                string                 `json:"modified_by,omitempty"`
	MostRecentJobSubmissionAt *time.Time             `json:"most_recent_job_submission_at,omitempty"`
	MostRecentJobSubmissionID string                 `json:"most_recent_job_submission_id,omitempty"`
	Name                      string                 `json:"name,omitempty" wbctl:"editable"`
	Nodepool                  string                 `json:"nodepool,omitempty" wbctl:"editable"`
	Properties                map[string]interface{} `json:"properties,omitempty" wbctl:"editable"`
	Schedule                  string                 `json:"schedule,omitempty" wbctl:"editable"`
	Services                  []ServiceRequest       `json:"services,omitempty" wbctl:"editable"`
	Spec                      map[string]interface{} `json:"spec,omitempty" wbctl:"editable"`
	SubPath                   string                 `json:"sub_path,omitempty" wbctl:"editable"`
	TenantID                  string                 `json:"tenant_id,omitempty"`
	Username                  string                 `json:"username,omitempty"`
	WorkspaceNames            map[string]string      `json:"workspace_names,omitempty" wbctl:"alias:volume_names"`
	Workspaces                []WorkspaceRequest     `json:"workspaces,omitempty" wbctl:"alias:volumes,editable"`
}

// Jobs is a slice of Job values.
type Jobs []Job

// AsJobSlice converts Jobs to a plain slice.
func AsJobSlice(source Jobs) []Job {
	return source
}

// JobSubmission represents a job submission record.
type JobSubmission struct {
	ID           string                      `json:"id,omitempty"`
	CreatedBy    string                      `json:"created_by,omitempty"`
	ModifiedBy   string                      `json:"modified_by,omitempty"`
	CreatedAt    *time.Time                  `json:"created_at,omitempty"`
	ModifiedAt   *time.Time                  `json:"modified_at,omitempty"`
	TenantID     string                      `json:"tenant_id,omitempty"`
	DataplaneID  string                      `json:"dataplane_id,omitempty"`
	JobID        string                      `json:"job_id,omitempty"`
	JobName      string                      `json:"job_name,omitempty"`
	Env          map[string]string           `json:"env,omitempty"`
	Workspaces   []WorkspaceRequest          `json:"workspaces,omitempty" wbctl:"alias:volumes"`
	Labels       map[string]string           `json:"labels,omitempty"`
	Nodepool     string                      `json:"nodepool,omitempty"`
	ControllerID string                      `json:"controller_id,omitempty"`
	URL          string                      `json:"url,omitempty"`
	CurrentState string                      `json:"current_state,omitempty"`
	Message      string                      `json:"message,omitempty"`
	PodStatus    []ExecuteableResourceStatus `json:"pod_status,omitempty"`
	Email        string                      `json:"email,omitempty"`
	Username     string                      `json:"username,omitempty"`
	CanceledAt   *time.Time                  `json:"canceled_at,omitempty"`
	CompletedAt  *time.Time                  `json:"completed_at,omitempty"`
	Runtime      float64                     `json:"runtime,omitempty"`
	Properties   map[string]interface{}      `json:"properties,omitempty"`
	Scheduled    bool                        `json:"scheduled,omitempty"`
	LogURL       string                      `json:"log_url,omitempty"`
}

// JobSubmissions is a slice of JobSubmission values.
type JobSubmissions []JobSubmission

// AsJobSubmissionSlice converts JobSubmissions to a plain slice.
func AsJobSubmissionSlice(source JobSubmissions) []JobSubmission {
	return source
}

// Nodepool represents a node pool.
type Nodepool struct {
	ID          string `json:"id,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	DataplaneID string `json:"dataplane_id,omitempty"`
	Description string `json:"description,omitempty"`
}

// Nodepools is a slice of Nodepool values.
type Nodepools []Nodepool

// AsNodepoolSlice converts JobSubmissions to a plain slice.
func AsNodepoolSlice(source JobSubmissions) []JobSubmission {
	return source
}

// PVCStatus describes a persistent volume claim status.
type PVCStatus struct {
	Deleted bool                               `json:"deleted,omitempty"`
	Status  corev1.PersistentVolumeClaimStatus `json:"status,omitempty"`
}

// Workspace represents a workspace.
type Workspace struct {
	ControllerID string                            `json:"controller_id,omitempty"`
	CreatedAt    *time.Time                        `json:"created_at,omitempty"`
	CreatedBy    string                            `json:"created_by,omitempty"`
	DataplaneID  string                            `json:"dataplane_id,omitempty" wbctl:"editable"`
	Description  string                            `json:"description,omitempty" wbctl:"editable"`
	Email        string                            `json:"email,omitempty"`
	ID           string                            `json:"id,omitempty"`
	Initializers []ServiceRequest                  `json:"initializers,omitempty" wbctl:"editable"`
	Message      string                            `json:"message,omitempty"`
	ModifiedAt   *time.Time                        `json:"modified_at,omitempty"`
	ModifiedBy   string                            `json:"modified_by,omitempty"`
	Name         string                            `json:"name,omitempty" wbctl:"editable"`
	PVCStatus    PVCStatus                         `json:"pvc_status,omitempty"`
	PodStatus    []ExecuteableResourceStatus       `json:"pod_status,omitempty"`
	PodStatusAt  *time.Time                        `json:"pod_status_at,omitempty"`
	Properties   map[string]interface{}            `json:"properties,omitempty" wbctl:"editable"`
	State        string                            `json:"state,omitempty"`
	StorageSizeG float64                           `json:"storage_size_g,omitempty"`
	StorageSpec  *corev1.PersistentVolumeClaimSpec `json:"storage_spec,omitempty" wbctl:"editable"`
	StorageType  string                            `json:"storage_type,omitempty" wbctl:"editable"`
	Username     string                            `json:"username,omitempty"`
}

// Workspaces is a slice of Workspace values.
type Workspaces []Workspace

// AsWorkspaceSlice converts Workspaces to a plain slice.
func AsWorkspaceSlice(source Workspaces) []Workspace {
	return source
}

// Workbench represents a workbench.
type Workbench struct {
	ID             string                      `json:"id,omitempty"`
	ApplicationID  string                      `json:"application_id,omitempty" wbctl:"editable"`
	ControllerID   string                      `json:"controller_id,omitempty"`
	CreatedAt      *time.Time                  `json:"created_at,omitempty"`
	CreatedBy      string                      `json:"created_by,omitempty"`
	CurrentState   string                      `json:"current_state,omitempty"`
	DataplaneID    string                      `json:"dataplane_id,omitempty" wbctl:"editable"`
	DesiredState   string                      `json:"desired_state,omitempty"` // On get
	Email          string                      `json:"email,omitempty"`
	Labels         map[string]string           `json:"labels,omitempty" wbctl:"editable"`
	Message        string                      `json:"message,omitempty"`
	ModifiedAt     *time.Time                  `json:"modified_at,omitempty"`
	ModifiedBy     string                      `json:"modified_by,omitempty"`
	Name           string                      `json:"name,omitempty" wbctl:"editable"`
	Nodepool       string                      `json:"nodepool,omitempty" wbctl:"editable"`
	PodStatus      []ExecuteableResourceStatus `json:"pod_status,omitempty"`
	PodStatusAt    *time.Time                  `json:"pod_status_at,omitempty"`
	Properties     map[string]interface{}      `json:"properties,omitempty" wbctl:"editable"`
	Services       []ServiceRequest            `json:"services,omitempty" wbctl:"editable"`
	State          string                      `json:"state,omitempty" wbctl:"editable,seedfield:DesiredState"`
	URL            string                      `json:"url,omitempty"`
	Username       string                      `json:"username,omitempty"`
	WorkspaceNames map[string]string           `json:"workspace_names,omitempty" wbctl:"alias:volume_names"`
	Workspaces     []WorkspaceRequest          `json:"workspaces,omitempty" wbctl:"alias:volumes,editable"`
}

// Workbenches is a slice of Workbench values.
type Workbenches []Workbench

// AsWorkbenchSlice converts Workbenches to a plain slice.
func AsWorkbenchSlice(source Workbenches) []Workbench {
	return source
}

// Application represents an application.
type Application struct {
	ID          string   `json:"id,omitempty"`
	Description string   `json:"description,omitempty"`
	DisplayName string   `json:"display_name,omitempty"`
	DataplaneID string   `json:"dataplane_id,omitempty"`
	Services    []string `json:"services,omitempty"`
}

// Applications is a slice of Application values.
type Applications []Application

// AsApplicationSlice converts Applications to a plain slice.
func AsApplicationSlice(source Applications) []Application {
	return source
}

// Service represents a service.
type Service struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	DataplaneID string `json:"dataplane_id,omitempty"`
}

// Services is a slice of Service values.
type Services []Service

// AsServiceSlice converts Services to a plain slice.
func AsServiceSlice(source Services) []Service {
	return source
}

// Initializer represents a workspace initializer.
type Initializer struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	DataplaneID string `json:"dataplane_id,omitempty"`
}

// Initializers is a slice of Initializer values.
type Initializers []Initializer

// AsInitializerSlice converts Initializers to a plain slice.
func AsInitializerSlice(source Initializers) []Initializer {
	return source
}

// ImageBuildLog represents an image build log entry.
type ImageBuildLog struct {
	BuildID       string     `json:"build_id,omitempty"`
	ContainerName string     `json:"container_name,omitempty"`
	ControllerID  string     `json:"controller_id,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
	CreatedBy     string     `json:"created_by,omitempty"`
	DataplaneID   string     `json:"dataplane_id,omitempty"`
	Log           string     `json:"log,omitempty"`
}

// ImageBuildLogs is a slice of ImageBuildLog values.
type ImageBuildLogs []ImageBuildLog

// AsImageBuildLogSlice converts ImageBuildLogs to a plain slice.
func AsImageBuildLogSlice(source ImageBuildLogs) []ImageBuildLog {
	return source
}

// ImageBuild represents an image build.
type ImageBuild struct {
	ID           string                      `json:"id,omitempty"`
	Name         string                      `json:"name,omitempty"`
	ControllerID string                      `json:"controller_id,omitempty"`
	CreatedAt    *time.Time                  `json:"created_at,omitempty"`
	CreatedBy    string                      `json:"created_by,omitempty"`
	DataplaneID  string                      `json:"dataplane_id,omitempty"`
	ImageID      string                      `json:"image_id,omitempty"`
	ModifiedBy   string                      `json:"modified_by,omitempty"`
	LogURL       string                      `json:"log_url,omitempty"`
	PodStatus    []ExecuteableResourceStatus `json:"pod_status,omitempty"`
}

// ImageBuilds is a slice of ImageBuild values.
type ImageBuilds []ImageBuild

// AsImageBuildSlice converts ImageBuilds to a plain slice.
func AsImageBuildSlice(source ImageBuilds) []ImageBuild {
	return source
}

// Image represents an image.
type Image struct {
	Context                      string                      `json:"context,omitempty" wbctl:"editable"`
	ControllerID                 string                      `json:"controller_id,omitempty"`
	CreatedAt                    *time.Time                  `json:"created_at,omitempty"`
	CreatedBy                    string                      `json:"created_by,omitempty"`
	CurrentState                 string                      `json:"current_state,omitempty"`
	DataplaneID                  string                      `json:"dataplane_id,omitempty" wbctl:"editable"`
	Description                  string                      `json:"description,omitempty" wbctl:"editable"`
	Email                        string                      `json:"email,omitempty"`
	ID                           string                      `json:"id,omitempty"`
	ImageName                    string                      `json:"image_name,omitempty" wbctl:"editable"`
	ImageRegistry                string                      `json:"image_registry,omitempty" wbctl:"editable"`
	ImageTag                     string                      `json:"image_tag,omitempty" wbctl:"editable"`
	Labels                       map[string]string           `json:"labels,omitempty" wbctl:"editable"`
	LatestImageBuildControllerID string                      `json:"latest_image_build_controller_id,omitempty"`
	LatestImageBuildID           string                      `json:"latest_image_build_id,omitempty"`
	Message                      string                      `json:"message,omitempty"`
	ModifiedAt                   *time.Time                  `json:"modified_at,omitempty"`
	ModifiedBy                   string                      `json:"modified_by,omitempty"`
	Name                         string                      `json:"name,omitempty" wbctl:"editable"`
	PodStatus                    []ExecuteableResourceStatus `json:"pod_status,omitempty"`
	Spec                         map[string]interface{}      `json:"spec,omitempty" wbctl:"editable"`
	SubPath                      string                      `json:"sub_path,omitempty" wbctl:"editable"`
	TenantID                     string                      `json:"tenant_id,omitempty"`
	Username                     string                      `json:"username,omitempty"`
}

// Images is a slice of Image values.
type Images []Image

// AsImageSlice converts Images to a plain slice.
func AsImageSlice(source Images) []Image {
	return source
}

// Dataplane represents a dataplane.
type Dataplane struct {
	ID          string                 `json:"id,omitempty"`
	CreatedBy   string                 `json:"created_by,omitempty"`
	ModifiedBy  string                 `json:"modified_by,omitempty"`
	CreatedAt   *time.Time             `json:"created_at,omitempty"`
	ModifiedAt  *time.Time             `json:"modified_at,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Spec        map[string]interface{} `json:"spec,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	State       string                 `json:"state,omitempty"`
	Message     string                 `json:"message,omitempty"`
}

// Dataplanes is a slice of Dataplane values.
type Dataplanes []Dataplane

// HomeDirAwareWorkspaceRequest wraps a workspace request with home directory metadata.
type HomeDirAwareWorkspaceRequest struct {
	WorkspaceRequest
	HomeDir bool
}

// StorageClass represents a storage class.
type StorageClass struct {
	ID          string                 `json:"id,omitempty"`
	DisplayName string                 `json:"display_name,omitempty"`
	Description string                 `json:"description,omitempty"`
	AccessMode  string                 `json:"access_mode,omitempty"`
	DataplaneID string                 `json:"dataplane_id,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

// StorageClasses is a slice of StorageClass values.
type StorageClasses []StorageClass

// Snapshot represents a snapshot.
type Snapshot struct {
	ControllerID    string                        `json:"controller_id,omitempty"`
	CreatedAt       *time.Time                    `json:"created_at,omitempty"`
	CreatedBy       string                        `json:"created_by,omitempty"`
	DataplaneID     string                        `json:"dataplane_id,omitempty" wbctl:"editable"`
	Description     string                        `json:"description,omitempty" wbctl:"editable"`
	Email           string                        `json:"email,omitempty"`
	ID              string                        `json:"id,omitempty"`
	Message         string                        `json:"message,omitempty"`
	ModifiedAt      *time.Time                    `json:"modified_at,omitempty"`
	ModifiedBy      string                        `json:"modified_by,omitempty"`
	Name            string                        `json:"name,omitempty" wbctl:"editable"`
	Properties      map[string]interface{}        `json:"properties,omitempty" wbctl:"editable"`
	Spec            *snapshot1.VolumeSnapshotSpec `json:"spec,omitempty" wbctl:"editable"`
	State           string                        `json:"state,omitempty"`
	TenantID        string                        `json:"tenant_id,omitempty"`
	Username        string                        `json:"username,omitempty"`
	VolumeStatus    *VolumeStatus                 `json:"volume_status,omitempty"`
	WorkspaceSource *Workspace                    `json:"workspace_source,omitempty" wbctl:"alias:volume_source"`
}

// VolumeStatus describes a snapshot volume status.
type VolumeStatus struct {
	Name    string                         `json:"name,omitempty"`
	Deleted bool                           `json:"deleted,omitempty"`
	Status  snapshot1.VolumeSnapshotStatus `json:"status,omitempty"`
}

// Snapshots is a slice of Snapshot values.
type Snapshots []Snapshot

// AsSnapshotSlice converts Snapshots to a plain slice.
func AsSnapshotSlice(source Snapshots) []Snapshot {
	return source
}

// Role represents a role.
type Role struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Auth0User is the representation of a user in the auth0Admin service.
type Auth0User struct {
	Version           int            `json:"version,omitempty"`
	ID                string         `json:"id,omitempty"`
	DatabaseID        string         `json:"databaseId,omitempty"`
	Name              string         `json:"name,omitempty"`
	Email             string         `json:"email,omitempty"`
	CreationTimeStamp time.Time      `json:"creationTimeStamp,omitempty"`
	Roles             []Role         `json:"roles,omitempty"`
	UserMetadata      map[string]any `json:"userMetadata,omitempty"`
}

// Auth0UserList is the representation of a list of users in the auth0Admin
// service.
type Auth0UserList struct {
	Version int         `json:"version"`
	Accept  string      `json:"accept"`
	Count   int         `json:"count"`
	Start   int         `json:"start"`
	Limit   int         `json:"limit"`
	Name    string      `json:"name"`
	Items   []Auth0User `json:"items"`
}

// Event represents an event.
type Event struct {
	ID           string     `json:"id,omitempty"`
	Action       string     `json:"action,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	Kind         string     `json:"kind,omitempty"`
	Note         string     `json:"note,omitempty"`
	Reason       string     `json:"reason,omitempty"`
	ResourceType string     `json:"resource_type,omitempty"`
	Type         string     `json:"type,omitempty"`
}

// Events is a slice of Event values.
type Events []Event

// AsEventSlice converts Events to a plain slice.
func AsEventSlice(source Events) []Event {
	return source
}

// Ping represents a ping response.
type Ping struct {
	DataplaneID string     `json:"dataplane_id,omitempty"`
	RequestAt   *time.Time `json:"request_at,omitempty"`
	ResponseAt  *time.Time `json:"status_updated_at,omitempty"`
	State       string     `json:"status_state,omitempty"`
	Message     string     `json:"status_message,omitempty"`
}

// Pings is a slice of Ping values.
type Pings []Ping

// AsPingSlice converts Pings to a plain slice.
func AsPingSlice(source Pings) []Ping {
	return source
}

// Conflict represents a conflict record.
type Conflict struct {
	ConflictControllerID  string `json:"conflict_controller_id,omitempty"`
	ConflictID            string `json:"conflict_id,omitempty"`
	ConflictName          string `json:"conflict_name,omitempty"`
	ConflictNodeID        string `json:"conflict_node_id,omitempty"`
	ConflictType          string `json:"conflict_type,omitempty"`
	ControllerID          string `json:"controller_id,omitempty"`
	ID                    string `json:"id,omitempty"`
	Name                  string `json:"name,omitempty"`
	NodeID                string `json:"node_id,omitempty"`
	Type                  string `json:"type,omitempty"`
	WorkspaceControllerID string `json:"workspace_controller_id,omitempty" wbctl:"alias:volume_controller_id"`
	WorkspaceID           string `json:"workspace_id,omitempty" wbctl:"alias:volume_id"`
	WorkspaceName         string `json:"workspace_name,omitempty" wbctl:"alias:volume_name"`
}

// Conflicts is a slice of Conflict values.
type Conflicts []Conflict

// AsConflictSlice converts Conflicts to a plain slice.
func AsConflictSlice(source Conflicts) []Conflict {
	return source
}

// Settings is a slice of Setting values.
type Settings map[string]any

// FunctionMetadata binds resource names to implementations.
type FunctionMetadata[T any] struct {
	Singular string
	Plural   string
	F        T
}

// SupportedResourceTypes returns the plural form of the resources represented
// by the given FunctionMetadata slice. The returned slice is sorted by the
// singular name so that error messages line up with the order of the commands
// which are sorted by the singular name.
func SupportedResourceTypes[T any](metadata []FunctionMetadata[T]) []string {
	slices.SortFunc(metadata, func(left, right FunctionMetadata[T]) int {
		return strings.Compare(left.Singular, right.Singular)
	})
	types := maputil.MapNoError(func(m FunctionMetadata[T]) string {
		if m.Plural != "" {
			return m.Plural
		}
		return m.Singular
	}, metadata)
	return types
}

// GetFunction returns the function bound to a resource type.
func GetFunction[T any](
	resourceType string,
	metadata []FunctionMetadata[T],
) (T, bool) {
	for _, gf := range metadata {
		if gf.Singular == resourceType || (gf.Plural != "" && gf.Plural == resourceType) {
			return gf.F, true
		}
	}
	var zero T
	return zero, false
}
