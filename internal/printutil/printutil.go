// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package printutil

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	snapshot1 "github.com/kubernetes-csi/external-snapshotter/client/v8/apis/volumesnapshot/v1"
	"github.com/sassoftware/sas-ggdk/pkg/pointer"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/timeutil"
	"github.com/sassoftware/wbctl/internal/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	tabwriterFlags    = 0
	tabwriterMinWidth = 6
	tabwriterPadChar  = ' '
	tabwriterPadding  = 3
	tabwriterTabWidth = 4
)

var marshalIndentedFunc = json.MarshalIndent

// AppendConflictDetails appends the description of the given types.Conflict as
// a tabbed string to the given writer.
func AppendConflictDetails(
	w io.Writer,
	resource types.Conflict,
) {
	AppendTabbedString(w, func(w io.Writer) {
		writeDetail(w, "", i18n.DescribeLabelID, resource.ID)
		writeDetail(w, "", i18n.DescribeLabelControllerID, resource.ControllerID)
		writeDetail(w, "", i18n.DescribeLabelName, resource.Name)
		writeDetail(w, "", i18n.DescribeLabelType, resource.Type)
		writeDetail(w, "", i18n.DescribeLabelConflictID, resource.ConflictID)
		writeDetail(w, "", i18n.DescribeLabelConflictControllerID, resource.ConflictControllerID)
		writeDetail(w, "", i18n.DescribeLabelConflictName, resource.ConflictName)
		writeDetail(w, "", i18n.DescribeLabelConflictType, resource.ConflictType)
		writeDetail(w, "", i18n.DescribeLabelWorkspaceID, resource.WorkspaceID)
		writeDetail(w, "", i18n.DescribeLabelWorkspaceControllerID, resource.WorkspaceControllerID)
		writeDetail(w, "", i18n.DescribeLabelWorkspaceName, resource.WorkspaceName)
	})
}

// AppendContainerStatuses appends the description of the given container
// statuses as a tabbed string to the given writer.
// nolint:cyclop
func AppendContainerStatuses(
	w io.Writer,
	title string,
	containerStatuses []corev1.ContainerStatus,
) {
	if len(containerStatuses) == 0 {
		return
	}
	_, _ = w.Write([]byte("  " + title + "\n"))
	for _, status := range containerStatuses {
		AppendTabbedString(w, func(w io.Writer) {
			_, _ = w.Write([]byte("    " + status.Name + ":\n"))
			writeDetail(w, "      ", i18n.DescribeLabelImage, status.Image)
			writeDetail(w, "      ", i18n.DescribeLabelReady, status.Ready)
			writeDetail(w, "      ", i18n.DescribeLabelStarted, pointer.UnPtr(status.Started))
			writeDetail(w, "      ", i18n.DescribeLabelRestartCount, status.RestartCount)
			writeDetail(w, "      ", i18n.DescribeLabelState, formatStateName(status.State))
		})
		AppendTabbedString(w, func(w io.Writer) {
			switch {
			case status.State.Waiting != nil:
				writeDetailIfNotEmpty(w, "        ", i18n.DescribeLabelReason, status.State.Waiting.Reason)
				writeDetailIfNotEmpty(w, "        ", i18n.DescribeLabelMessage, status.State.Waiting.Message)
			case status.State.Terminated != nil:
				writeDetailIfNotEmpty(w, "        ", i18n.DescribeLabelReason, status.State.Terminated.Reason)
				writeDetailIfNotEmpty(w, "        ", i18n.DescribeLabelMessage, status.State.Terminated.Message)
				writeDetailIfNotEmpty(w, "        ", i18n.DescribeLabelExitCode, status.State.Terminated.ExitCode)
				writeDetailIfNotEmpty(
					w,
					"        ",
					i18n.DescribeLabelStartedAt,
					metaTimePtrToString(&status.State.Terminated.StartedAt),
				)
				writeDetailIfNotEmpty(
					w,
					"        ",
					i18n.DescribeLabelFinishedAt,
					metaTimePtrToString(&status.State.Terminated.FinishedAt),
				)
			case status.State.Running != nil:
				writeDetail(
					w,
					"        ",
					i18n.DescribeLabelStartedAt,
					status.State.Running.StartedAt.Format(time.RFC3339Nano),
				)
			}
		})
		if status.Resources != nil && len(status.Resources.Limits) > 0 {
			AppendTabbedString(w, func(w io.Writer) {
				_, _ = fmt.Fprintf(w, "      %s\n", i18n.Text(i18n.DescribeLabelLimits))
				writeDetailIfNotEmpty(w, "        ", i18n.DescribeLabelCPU, status.Resources.Limits.Cpu().String())
				writeDetailIfNotEmpty(w, "        ", i18n.DescribeLabelMemory, status.Resources.Limits.Memory().String())
			})
		}
		if status.Resources != nil && len(status.Resources.Requests) > 0 {
			AppendTabbedString(w, func(w io.Writer) {
				_, _ = fmt.Fprintf(w, "      %s\n", i18n.Text(i18n.DescribeLabelRequests))
				writeDetailIfNotEmpty(w, "        ", i18n.DescribeLabelCPU, status.Resources.Requests.Cpu().String())
				writeDetailIfNotEmpty(w, "        ", i18n.DescribeLabelMemory, status.Resources.Requests.Memory().String())
			})
		}
		if len(status.AllocatedResources) > 0 {
			AppendTabbedString(w, func(w io.Writer) {
				_, _ = fmt.Fprintf(w, "      %s\n", i18n.Text(i18n.DescribeLabelAllocated))
				writeDetailIfNotEmpty(w, "        ", i18n.DescribeLabelCPU, status.AllocatedResources.Cpu())
				writeDetailIfNotEmpty(w, "        ", i18n.DescribeLabelMemory, status.AllocatedResources.Memory())
			})
		}
		if len(status.VolumeMounts) > 0 {
			_, _ = fmt.Fprintf(w, "      %s\n", i18n.Text(i18n.DescribeLabelVolumeMounts))
			mounts := []string{}
			for _, mount := range status.VolumeMounts {
				perms := ""
				if mount.ReadOnly {
					perms = i18n.Text(i18n.DescribeReadOnly)
				}
				mounts = append(mounts, fmt.Sprintf("        %s %s", mount.MountPath, perms))
			}
			slices.Sort(mounts)
			_, _ = w.Write([]byte(strings.Join(mounts, "\n") + "\n"))
		}
		if status.User != nil && status.User.Linux != nil {
			_, _ = fmt.Fprintf(w, "      %s\n", i18n.Text(i18n.DescribeLabelLinuxUser))
			AppendTabbedString(w, func(w io.Writer) {
				writeDetail(w, "        ", i18n.DescribeLabelUID, status.User.Linux.UID)
				writeDetail(w, "        ", i18n.DescribeLabelGID, status.User.Linux.GID)
				if len(status.User.Linux.SupplementalGroups) > 0 {
					gids := make([]string, 0, len(status.User.Linux.SupplementalGroups))
					for _, gid := range status.User.Linux.SupplementalGroups {
						gids = append(gids, strconv.FormatInt(gid, 10))
					}
					writeDetail(w, "        ", i18n.DescribeLabelAdditionalGIDs, strings.Join(gids, ", "))
				}
			})
		}
	}
}

// AppendDataplaneDetails appends the description of the given types.Dataplane
// as a tabbed string to the given writer.
func AppendDataplaneDetails(
	w io.Writer,
	resource types.Dataplane,
) {
	AppendTabbedString(w, func(w io.Writer) {
		writeDetail(w, "", i18n.DescribeLabelID, resource.ID)
		writeDetail(w, "", i18n.DescribeLabelName, resource.Name)
		writeDetailIfNotEmpty(w, "", i18n.DescribeLabelDescription, resource.Description)
		writeDetail(w, "", i18n.DescribeLabelCreatedBy, resource.CreatedBy)
		writeDetail(w, "", i18n.DescribeLabelCreatedAt, TimePtrToStringWithAge(resource.CreatedAt))
		writeDetail(w, "", i18n.DescribeLabelModifiedAt, TimePtrToStringWithAge(resource.ModifiedAt))
		writeDetail(w, "", i18n.DescribeLabelState, resource.State)
		writeDetailBlockIfNotEmpty(w, "", i18n.DescribeLabelSpec, specAsString(resource.Spec))
		writeDetailIfNotEmpty(w, "", i18n.DescribeLabelMessage, resource.Message)
	})
}

// AppendEnvDetails appends the description of the given environment as a tabbed
// string to the given writer.
func AppendEnvDetails(
	w io.Writer,
	env map[string]string,
) {
	if len(env) > 0 {
		_, _ = fmt.Fprintf(w, "%s\n", i18n.Text(i18n.DescribeLabelEnvironment))
		keys := make([]string, 0, len(env))
		for k := range env {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			AppendTabbedString(w, func(w io.Writer) {
				_, _ = fmt.Fprintf(w, "  %s:\t%s\n", k, env[k])
			})
		}
	}
}

// AppendImageBuildDetails appends the description of the given types.ImageBuild
// as a tabbed string to the given writer.
func AppendImageBuildDetails(
	w io.Writer,
	resource types.ImageBuild,
) {
	AppendTabbedString(w, func(w io.Writer) {
		writeDetail(w, "", i18n.DescribeLabelID, resource.ID)
		writeDetail(w, "", i18n.DescribeLabelControllerID, resource.ControllerID)
		writeDetail(w, "", i18n.DescribeLabelName, resource.Name)
		writeDetail(w, "", i18n.DescribeLabelCreatedBy, resource.CreatedBy)
		writeDetail(w, "", i18n.DescribeLabelCreatedAt, TimePtrToStringWithAge(resource.CreatedAt))
		writeDetail(w, "", i18n.DescribeLabelImageID, resource.ImageID)
	})
}

// AppendImageDetails appends the description of the given types.Image as a
// tabbed string to the given writer.
func AppendImageDetails(
	w io.Writer,
	resource types.Image,
) {
	image := resource.ImageRegistry + "/" + resource.ImageName
	if resource.ImageTag != "" {
		image += ":" + resource.ImageTag
	}
	AppendTabbedString(w, func(w io.Writer) {
		writeDetail(w, "", i18n.DescribeLabelID, resource.ID)
		writeDetail(w, "", i18n.DescribeLabelControllerID, resource.ControllerID)
		writeDetail(w, "", i18n.DescribeLabelName, resource.Name)
		writeDetailIfNotEmpty(w, "", i18n.DescribeLabelDescription, resource.Description)
		writeDetail(w, "", i18n.DescribeLabelCreatedBy, resource.Username+" ("+resource.Email+")")
		writeDetail(w, "", i18n.DescribeLabelCreatedAt, TimePtrToStringWithAge(resource.CreatedAt))
		writeDetail(w, "", i18n.DescribeLabelModifiedAt, TimePtrToStringWithAge(resource.ModifiedAt))
		writeDetail(w, "", i18n.DescribeLabelImage, image)
		writeDetailIfNotEmpty(w, "", i18n.DescribeLabelContext, resource.Context)
		writeDetailBlockIfNotEmpty(w, "", i18n.DescribeLabelSpec, specAsString(resource.Spec))
		writeDetailIfNotEmpty(w, "", i18n.DescribeLabelSubPath, resource.SubPath)
		writeDetailIfNotEmpty(w, "", i18n.DescribeLabelLatestImageBuildID, resource.LatestImageBuildID)
		if resource.CurrentState != "" {
			_, _ = fmt.Fprintf(w, "%s\t%s\n", i18n.Text(i18n.DescribeLabelState), resource.CurrentState)
		}
		if resource.Message != "" {
			_, _ = fmt.Fprintf(w, "%s\t%s\n", i18n.Text(i18n.DescribeLabelMessage), resource.Message)
		}
	})
}

// AppendJobDetails appends the description of the given types.Job as a tabbed
// string to the given writer.
func AppendJobDetails(
	w io.Writer,
	resource types.Job,
) {
	AppendTabbedString(w, func(w io.Writer) {
		writeDetail(w, "", i18n.DescribeLabelID, resource.ID)
		writeDetail(w, "", i18n.DescribeLabelName, resource.Name)
		writeDetailIfNotEmpty(w, "", i18n.DescribeLabelDescription, resource.Description)
		writeDetail(w, "", i18n.DescribeLabelCreatedBy, resource.Username+" ("+resource.Email+")")
		writeDetail(w, "", i18n.DescribeLabelCreatedAt, TimePtrToStringWithAge(resource.CreatedAt))
		writeDetail(w, "", i18n.DescribeLabelModifiedAt, TimePtrToStringWithAge(resource.ModifiedAt))
		if resource.MostRecentJobSubmissionAt != nil {
			writeDetailIfNotEmpty(
				w,
				"",
				i18n.DescribeLabelMostRecentJobRunAt,
				resource.MostRecentJobSubmissionAt.Format(time.RFC3339Nano),
			)
		}
		writeDetailIfNotEmpty(w, "", i18n.DescribeLabelMostRecentJobRunID, resource.MostRecentJobSubmissionID)
		writeDetail(w, "", i18n.DescribeLabelNodepool, resource.Nodepool)
		writeDetailIfNotEmpty(w, "", i18n.DescribeLabelApplication, resource.ApplicationID)
		writeDetailIfNotEmpty(w, "", i18n.DescribeLabelContext, resource.Context)
		writeDetailBlockIfNotEmpty(w, "", i18n.DescribeLabelSpec, specAsString(resource.Spec))
		writeDetailIfNotEmpty(w, "", i18n.DescribeLabelSubPath, resource.SubPath)
		writeDetailIfNotEmpty(w, "", i18n.DescribeLabelSchedule, resource.Schedule)
	})
}

// AppendJobSubmissionDetails appends the description of the given
// types.JobSubmission as a tabbed string to the given writer.
func AppendJobSubmissionDetails(
	w io.Writer,
	resource types.JobSubmission,
) {
	AppendTabbedString(w, func(w io.Writer) {
		writeDetail(w, "", i18n.DescribeLabelID, resource.ID)
		writeDetail(w, "", i18n.DescribeLabelControllerID, resource.ControllerID)
		writeDetail(w, "", i18n.DescribeLabelJobName, resource.JobName)
		writeDetail(w, "", i18n.DescribeLabelJobID, resource.JobID)
		writeDetail(w, "", i18n.DescribeLabelURL, resource.URL)
		writeDetail(w, "", i18n.DescribeLabelLogURL, resource.LogURL)
		writeDetail(w, "", i18n.DescribeLabelCreatedBy, resource.Username+" ("+resource.Email+")")
		writeDetail(w, "", i18n.DescribeLabelCreatedAt, TimePtrToStringWithAge(resource.CreatedAt))
		writeDetail(w, "", i18n.DescribeLabelModifiedAt, TimePtrToStringWithAge(resource.ModifiedAt))
		if resource.CanceledAt != nil {
			writeDetail(w, "", i18n.DescribeLabelCanceledAt, TimePtrToStringWithAge(resource.CanceledAt))
		}
		if resource.CompletedAt != nil {
			writeDetail(w, "", i18n.DescribeLabelCompletedAt, TimePtrToStringWithAge(resource.CompletedAt))
		}
		_, _ = fmt.Fprintf(w, "%s\t%f\n", i18n.Text(i18n.DescribeLabelRuntime), resource.Runtime)
		writeDetailIfNotEmpty(w, "", i18n.DescribeLabelNodepool, resource.Nodepool)
		writeDetail(w, "", i18n.DescribeLabelScheduled, resource.Scheduled)
		writeDetail(w, "", i18n.DescribeLabelState, resource.CurrentState)
		writeDetailIfNotEmpty(w, "", i18n.DescribeLabelMessage, resource.Message)
	})
}

// AppendLabelDetails appends the description of the given labels as a tabbed
// string to the given writer.
func AppendLabelDetails(
	w io.Writer,
	labels map[string]string,
) {
	if len(labels) == 0 {
		return
	}
	AppendTabbedString(w, func(w io.Writer) {
		_, _ = fmt.Fprintf(w, "%s\n", i18n.Text(i18n.DescribeLabelLabels))
		keys := make([]string, 0, len(labels))
		for k := range labels {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		for _, k := range keys {
			_, _ = fmt.Fprintf(w, "  %s:\t%s\n", k, labels[k])
		}
	})
}

// AppendPodStatusDetails appends the description of the given pod statuses as a
// tabbed string to the given writer.
func AppendPodStatusDetails(
	w io.Writer,
	podStatuses []types.ExecuteableResourceStatus,
) {
	if len(podStatuses) > 0 {
		for _, ps := range podStatuses {
			AppendTabbedString(w, func(w io.Writer) {
				if ps.Name != "" {
					title := i18n.Format(i18n.DescribePodHeading, i18n.A{constants.I18nArgKeyName: ps.Name})
					_, _ = fmt.Fprintf(w, "%s:\n", title)
				}
				writeDetailIfNotEmpty(w, "  ", i18n.DescribeLabelPhase, string(ps.Status.Phase))
				if ps.Deleted {
					writeDetail(w, "  ", i18n.DescribeLabelDeleted, ps.Deleted)
				}
				writeDetailIfNotEmpty(w, "  ", i18n.DescribeLabelMessage, ps.Status.Message)
				writeDetailIfNotEmpty(w, "  ", i18n.DescribeLabelReason, ps.Status.Reason)
				writeDetailIfNotEmpty(w, "  ", i18n.DescribeLabelNominatedNode, ps.Status.NominatedNodeName)
				writeDetailIfNotEmpty(w, "  ", i18n.DescribeLabelIPs, formatIPs(ps.Status.PodIP, ps.Status.PodIPs))
			})
			AppendTabbedString(w, func(w io.Writer) {
				if len(ps.Status.Conditions) > 0 {
					_, _ = fmt.Fprintf(w, "  %s\n", i18n.Text(i18n.DescribeLabelConditions))
					for _, cond := range ps.Status.Conditions {
						_, _ = fmt.Fprintf(w, "    %s\t%s\n", cond.Type, cond.Status)
					}
				}
			})
			AppendContainerStatuses(w, i18n.Text(i18n.DescribeLabelInitContainers), ps.Status.InitContainerStatuses)
			AppendContainerStatuses(w, i18n.Text(i18n.DescribeLabelContainers), ps.Status.ContainerStatuses)
			AppendContainerStatuses(
				w,
				i18n.Text(i18n.DescribeLabelEphemeralContainers),
				ps.Status.EphemeralContainerStatuses,
			)
		}
	}
}

// AppendPropertiesDetails appends the description of the given properties as a
// tabbed string to the given writer.
func AppendPropertiesDetails(
	w io.Writer,
	properties map[string]interface{},
) {
	if len(properties) == 0 {
		return
	}
	_, _ = fmt.Fprintf(w, "%s\n", i18n.Text(i18n.DescribeLabelProperties))
	AppendTabbedString(w, func(w io.Writer) {
		keys := make([]string, 0, len(properties))
		for k := range properties {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			_, _ = fmt.Fprintf(w, "  %s:\t%v\n", k, properties[k])
		}
	})
}

// AppendRolesDetails appends the description of the given types.Role slice as a
// tabbed string to the given writer.
func AppendRolesDetails(
	w io.Writer,
	roles []types.Role,
) {
	if len(roles) > 0 {
		_, _ = fmt.Fprintf(w, "%s\n", i18n.Text(i18n.DescribeLabelRoles))
		AppendTabbedString(w, func(w io.Writer) {
			slices.SortFunc(roles, func(a, b types.Role) int {
				return strings.Compare(a.Name, b.Name)
			})
			for _, role := range roles {
				_, _ = fmt.Fprintf(w, "  - %s\n", role.Name)
			}
		})
	}
}

// AppendServiceRequestsDetails appends a service section with its localized
// heading.
func AppendServiceRequestsDetails(w io.Writer, services []types.ServiceRequest) {
	appendServiceRequestDetails(w, i18n.DescribeLabelServices, services)
}

// AppendInitializerRequestsDetails appends an initializer section with its
// localized heading.
func AppendInitializerRequestsDetails(w io.Writer, services []types.ServiceRequest) {
	appendServiceRequestDetails(w, i18n.DescribeLabelInitializers, services)
}

func appendServiceRequestDetails(
	w io.Writer,
	headingID string,
	services []types.ServiceRequest,
) {
	if len(services) > 0 {
		_, _ = fmt.Fprintf(w, "%s\n", i18n.Text(headingID))
		for i, svc := range services {
			AppendTabbedString(w, func(w io.Writer) {
				_, _ = fmt.Fprintf(w, "  %d:\n", i)
				if svc.ID != "" {
					writeDetail(w, "    ", i18n.DescribeLabelID, svc.ID)
				}
				if svc.Image != "" {
					writeDetail(w, "    ", i18n.DescribeLabelImage, svc.Image)
				}
				if len(svc.Env) > 0 {
					_, _ = fmt.Fprintf(w, "    %s\n", i18n.Text(i18n.DescribeLabelEnvironment))
					for k, v := range svc.Env {
						_, _ = fmt.Fprintf(w, "      %s:\t%s\n", k, v)
					}
				}
			})
		}
	}
}

// AppendMetadataDetails appends the description of the given settings as a
// tabbed string to the given writer.
func AppendMetadataDetails(
	w io.Writer,
	settings map[string]any,
) {
	if len(settings) > 0 {
		_, _ = fmt.Fprintf(w, "%s\n", i18n.Text(i18n.DescribeLabelMetadata))
		AppendTabbedString(w, func(w io.Writer) {
			keys := make([]string, 0, len(settings))
			for k := range settings {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, key := range keys {
				_, _ = fmt.Fprintf(w, "  %s:\t%v\n", key, settings[key])
			}
		})
	}
}

// AppendSnapshotDetails appends the description of the given types.Snapshot as
// a tabbed string to the given writer.
func AppendSnapshotDetails(
	w io.Writer,
	resource types.Snapshot,
) {
	AppendTabbedString(w, func(w io.Writer) {
		writeDetail(w, "", i18n.DescribeLabelID, resource.ID)
		writeDetail(w, "", i18n.DescribeLabelControllerID, resource.ControllerID)
		writeDetail(w, "", i18n.DescribeLabelName, resource.Name)
		writeDetail(w, "", i18n.DescribeLabelDescription, resource.Description)
		writeDetail(w, "", i18n.DescribeLabelCreatedBy, resource.Username+" ("+resource.Email+")")
		writeDetail(w, "", i18n.DescribeLabelCreatedAt, TimePtrToStringWithAge(resource.CreatedAt))
		writeDetail(w, "", i18n.DescribeLabelModifiedAt, TimePtrToStringWithAge(resource.ModifiedAt))
		writeDetailBlockIfNotEmpty(w, "", i18n.DescribeLabelSpec, snapshotSpecAsString(resource.Spec))
		writeDetailIfNotEmpty(w, "", i18n.DescribeLabelMessage, resource.Message)
	})
}

// AppendStorageSpecDetails appends the description of the given
// corev1.PersistentVolumeClaimSpec as a tabbed string to the given writer.
func AppendStorageSpecDetails(
	w io.Writer,
	spec *corev1.PersistentVolumeClaimSpec,
	prefix string,
) {
	if spec == nil {
		return
	}
	accessModes := make([]string, 0, len(spec.AccessModes))
	for _, mode := range spec.AccessModes {
		accessModes = append(accessModes, string(mode))
	}
	AppendTabbedString(w, func(w io.Writer) {
		writeDetailIfNotEmpty(w, prefix, i18n.DescribeLabelAccessModes, strings.Join(accessModes, ", "))
		writeDetailIfNotEmpty(w, prefix, i18n.DescribeLabelStorageClass, pointer.UnPtr(spec.StorageClassName))
		writeDetailIfNotEmpty(w, prefix, i18n.DescribeLabelSize, spec.Resources.Requests.Storage().String())
	})
}

// AppendTabbedString appends the given functions output with tabwriter formatting.
func AppendTabbedString(
	buffer io.Writer,
	f func(io.Writer),
) {
	w := tabwriter.NewWriter(
		buffer,
		tabwriterMinWidth,
		tabwriterTabWidth,
		tabwriterPadding,
		tabwriterPadChar,
		tabwriterFlags,
	)
	f(w)
	_ = w.Flush()
}

// AppendUserDetails appends the description of the given types.User as a tabbed
// string to the given writer.
func AppendUserDetails(
	w io.Writer,
	user types.Auth0User,
) {
	AppendTabbedString(w, func(w io.Writer) {
		writeDetail(w, "", i18n.DescribeLabelID, user.ID)
		writeDetail(w, "", i18n.DescribeLabelDatabaseID, user.DatabaseID)
		writeDetail(w, "", i18n.DescribeLabelCreatedAt, TimePtrToStringWithAge(&user.CreationTimeStamp))
		writeDetail(w, "", i18n.DescribeLabelUsername, user.Name)
		writeDetail(w, "", i18n.DescribeLabelEmail, user.Email)
	})
}

// AppendVolumeDetails appends the description of the given types.VolumeStatus
// as a tabbed string to the given writer.
func AppendVolumeDetails(
	w io.Writer,
	volumeStatus *types.VolumeStatus,
) {
	_, _ = fmt.Fprintf(w, "%s\n", i18n.Text(i18n.DescribeLabelVolumeStatus))
	AppendTabbedString(w, func(w io.Writer) {
		if volumeStatus != nil {
			restoreSize := i18n.Text(i18n.DescribeUnknown)
			if size := volumeStatus.Status.RestoreSize; size != nil {
				restoreSize = size.String()
			}
			writeDetail(w, "  ", i18n.DescribeLabelReadyToUse, pointer.UnPtr(volumeStatus.Status.ReadyToUse))
			writeDetail(w, "  ", i18n.DescribeLabelRestoreSize, restoreSize)
			writeDetail(
				w,
				"  ",
				i18n.DescribeLabelContentName,
				pointer.UnPtr(volumeStatus.Status.BoundVolumeSnapshotContentName),
			)
			writeDetail(w, "  ", i18n.DescribeLabelCreationTime, metaTimePtrToString(volumeStatus.Status.CreationTime))
		} else {
			unknown := i18n.Text(i18n.DescribeUnknown)
			writeDetail(w, "  ", i18n.DescribeLabelReadyToUse, false)
			writeDetail(w, "  ", i18n.DescribeLabelRestoreSize, unknown)
			writeDetail(w, "  ", i18n.DescribeLabelContentName, unknown)
			writeDetail(w, "  ", i18n.DescribeLabelCreationTime, unknown)
		}
	})
}

// AppendWorkbenchDetails appends the description of the given types.Workbench
// as a tabbed string to the given writer.
func AppendWorkbenchDetails(
	w io.Writer,
	resource types.Workbench,
) {
	AppendTabbedString(w, func(w io.Writer) {
		writeDetail(w, "", i18n.DescribeLabelID, resource.ID)
		writeDetail(w, "", i18n.DescribeLabelControllerID, resource.ControllerID)
		writeDetail(w, "", i18n.DescribeLabelName, resource.Name)
		writeDetail(w, "", i18n.DescribeLabelURL, resource.URL)
		writeDetail(w, "", i18n.DescribeLabelCreatedBy, resource.Username+" ("+resource.Email+")")
		writeDetail(w, "", i18n.DescribeLabelCreatedAt, TimePtrToStringWithAge(resource.CreatedAt))
		writeDetail(w, "", i18n.DescribeLabelModifiedAt, TimePtrToStringWithAge(resource.ModifiedAt))
		writeDetail(w, "", i18n.DescribeLabelNodepool, resource.Nodepool)
		writeDetail(w, "", i18n.DescribeLabelDesiredState, resource.DesiredState)
		writeDetail(w, "", i18n.DescribeLabelCurrentState, resource.CurrentState)
		writeDetailIfNotEmpty(w, "", i18n.DescribeLabelMessage, resource.Message)
		writeDetailIfNotEmpty(w, "", i18n.DescribeLabelApplication, resource.ApplicationID)
	})
}

// AppendWorkspaceDetails appends the description of the given types.Workspace
// as a tabbed string to the given writer.
func AppendWorkspaceDetails(
	w io.Writer,
	resource types.Workspace,
) {
	AppendTabbedString(w, func(w io.Writer) {
		writeDetail(w, "", i18n.DescribeLabelID, resource.ID)
		writeDetail(w, "", i18n.DescribeLabelControllerID, resource.ControllerID)
		writeDetail(w, "", i18n.DescribeLabelName, resource.Name)
		writeDetailIfNotEmpty(w, "", i18n.DescribeLabelDescription, resource.Description)
		writeDetail(w, "", i18n.DescribeLabelStorageType, resource.StorageType)
		writeDetail(w, "", i18n.DescribeLabelCreatedBy, resource.Username+" ("+resource.Email+")")
		writeDetail(w, "", i18n.DescribeLabelCreatedAt, TimePtrToStringWithAge(resource.CreatedAt))
		writeDetail(w, "", i18n.DescribeLabelModifiedAt, TimePtrToStringWithAge(resource.ModifiedAt))
		writeDetail(w, "", i18n.DescribeLabelState, resource.State)
		writeDetailIfNotEmpty(w, "", i18n.DescribeLabelMessage, resource.Message)
		writeDetailIfNotEmpty(w, "", i18n.DescribeLabelPhase, resource.PVCStatus.Status.Phase)
	})
}

// AppendWorkspaceRequestsDetails appends the description of the given
// types.WorkspaceRequest slice as a tabbed string to the given writer.
func AppendWorkspaceRequestsDetails(
	w io.Writer,
	workspaces []types.WorkspaceRequest,
) {
	if len(workspaces) > 0 {
		_, _ = fmt.Fprintf(w, "%s\n", i18n.Text(i18n.DescribeLabelVolumes))
		for i, wksp := range workspaces {
			AppendTabbedString(w, func(w io.Writer) {
				_, _ = fmt.Fprintf(w, "  %d:\n", i)
				if wksp.ID != "" {
					writeDetail(w, "    ", i18n.DescribeLabelID, wksp.ID)
				}
				if wksp.MountName != "" {
					writeDetail(w, "    ", i18n.DescribeLabelMountName, wksp.MountName)
				}
				if wksp.EphemeralPVC != nil {
					_, _ = fmt.Fprintf(w, "    %s\n", i18n.Text(i18n.DescribeLabelEphemeralPVC))
					AppendStorageSpecDetails(w, wksp.EphemeralPVC, "      ")
				}
			})
		}
	}
}

// QuantityPtrToString returns a string representation of the given
// resource.Quantity pointer, or a placeholder if the pointer is nil.
func QuantityPtrToString(size *resource.Quantity) string {
	if size == nil {
		return constants.ValueUnknown
	}
	return size.String()
}

// SetMarshalIndentedFunc overrides the marshaling function used by this
// package.
func SetMarshalIndentedFunc(
	f func(
		any,
		string,
		string,
	) ([]byte, error),
) {
	marshalIndentedFunc = f
}

// TimePtrToString returns a string representation of the given time.Time
// pointer, or a placeholder if the pointer is nil.
func TimePtrToString(t *time.Time) string {
	if t == nil || t.IsZero() {
		return constants.ValueUnknown
	}
	return t.UTC().Format(time.RFC3339)
}

// TimePtrToStringWithAge returns a string representation of the given time.Time
// pointer with a human friendly age, or a placeholder if the pointer is nil.
func TimePtrToStringWithAge(t *time.Time) string {
	if t == nil || t.IsZero() {
		return i18n.Text(i18n.DescribeUnknown)
	}
	return i18n.Format(i18n.DescribeTimestampWithAge, i18n.A{
		"timestamp": t.UTC().Format(time.RFC3339),
		"age":       timeutil.Since(t),
	})
}

func formatIPs(
	podIP string,
	podIPs []corev1.PodIP,
) string {
	if len(podIPs) > 0 {
		ips := make([]string, 0, len(podIPs))
		for _, pip := range podIPs {
			ips = append(ips, pip.IP)
		}
		return strings.Join(ips, ", ")
	}
	return podIP
}

func formatStateName(state corev1.ContainerState) string {
	switch {
	case state.Waiting != nil:
		return i18n.Text(i18n.DescribeContainerWaiting)
	case state.Terminated != nil:
		return i18n.Text(i18n.DescribeContainerTerminated)
	case state.Running != nil:
		return i18n.Text(i18n.DescribeContainerRunning)
	default:
		return i18n.Text(i18n.DescribeUnknown)
	}
}

func marshalIndented(
	v any,
	prefix string,
) (string, error) {
	r, err := marshalIndentedFunc(v, prefix, "  ")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(r)), nil
}

func metaTimePtrToString(t *metav1.Time) string {
	if t == nil || t.IsZero() {
		return i18n.Text(i18n.DescribeUnknown)
	}
	return TimePtrToStringWithAge(&t.Time)
}

func snapshotSpecAsString(v *snapshot1.VolumeSnapshotSpec) string {
	if v == nil {
		return ""
	}
	r, err := marshalIndented(v, "")
	if err != nil {
		return i18n.Format(i18n.DescribeSpecError, i18n.A{constants.I18nArgKeyError: err.Error()})
	}
	return r
}

func specAsString(v map[string]interface{}) string {
	if len(v) == 0 {
		return ""
	}
	r, err := marshalIndented(v, "")
	if err != nil {
		return i18n.Format(i18n.DescribeSpecError, i18n.A{constants.I18nArgKeyError: err.Error()})
	}
	return strings.TrimSpace(r)
}

// writeDetail writes a labeled scalar value as a tab-separated row. The
// translated label includes its separator punctuation.
func writeDetail(w io.Writer, prefix, label string, value any) {
	_, _ = fmt.Fprintf(w, "%s%s\t%v\n", prefix, i18n.Text(label), value)
}

// writeDetailBlock writes a labeled multiline value on the lines below its label.
func writeDetailBlock(w io.Writer, prefix, label, value string) {
	if !strings.Contains(value, "\n") {
		writeDetail(w, prefix, label, value)
		return
	}

	_, _ = fmt.Fprintf(w, "%s%s\n", prefix, i18n.Text(label))
	for _, line := range strings.Split(strings.TrimSuffix(value, "\n"), "\n") {
		_, _ = fmt.Fprintf(w, "%s  %s\n", prefix, line)
	}
}

func writeDetailIfNotEmpty(w io.Writer, prefix, label string, value any) {
	if value != "" {
		writeDetail(w, prefix, label, value)
	}
}

func writeDetailBlockIfNotEmpty(w io.Writer, prefix, label, value string) {
	if value != "" {
		writeDetailBlock(w, prefix, label, value)
	}
}
