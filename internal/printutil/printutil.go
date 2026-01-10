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
		_, _ = fmt.Fprintf(w, "ID:\t%s\n", resource.ID)
		_, _ = fmt.Fprintf(w, "Controller ID:\t%s\n", resource.ControllerID)
		_, _ = fmt.Fprintf(w, "Name:\t%s\n", resource.Name)
		_, _ = fmt.Fprintf(w, "Type:\t%s\n", resource.Type)
		_, _ = fmt.Fprintf(w, "Conflict ID:\t%s\n", resource.ConflictID)
		_, _ = fmt.Fprintf(w, "Conflict Controller ID:\t%s\n", resource.ConflictControllerID)
		_, _ = fmt.Fprintf(w, "Conflict Name:\t%s\n", resource.ConflictName)
		_, _ = fmt.Fprintf(w, "Conflict Type:\t%s\n", resource.ConflictType)
		_, _ = fmt.Fprintf(w, "Workspace ID:\t%s\n", resource.WorkspaceID)
		_, _ = fmt.Fprintf(w, "Workspace Controller ID:\t%s\n", resource.WorkspaceControllerID)
		_, _ = fmt.Fprintf(w, "Workspace Name:\t%s\n", resource.WorkspaceName)
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
			_, _ = fmt.Fprintf(w, "      Image:\t%s\n", status.Image)
			_, _ = fmt.Fprintf(w, "      Ready:\t%t\n", status.Ready)
			_, _ = fmt.Fprintf(w, "      Started:\t%t\n", pointer.UnPtr(status.Started))
			_, _ = fmt.Fprintf(w, "      Restart Count:\t%d\n", status.RestartCount)
			_, _ = fmt.Fprintf(w, "      State:\t%s\n", formatStateName(status.State))
		})
		AppendTabbedString(w, func(w io.Writer) {
			switch {
			case status.State.Waiting != nil:
				writeIfNotEmptyf(w, "        Reason:\t%s\n", status.State.Waiting.Reason)
				writeIfNotEmptyf(w, "        Message:\t%s\n", status.State.Waiting.Message)
			case status.State.Terminated != nil:
				writeIfNotEmptyf(w, "        Reason:\t%s\n", status.State.Terminated.Reason)
				writeIfNotEmptyf(w, "        Message:\t%s\n", status.State.Terminated.Message)
				writeIfNotEmptyf(w, "        Exit Code:\t%d\n", status.State.Terminated.ExitCode)
				writeIfNotEmptyf(
					w,
					"        Started At:\t%s\n",
					metaTimePtrToString(&status.State.Terminated.StartedAt),
				)
				writeIfNotEmptyf(
					w,
					"        Finished At:\t%s\n",
					metaTimePtrToString(&status.State.Terminated.FinishedAt),
				)
			case status.State.Running != nil:
				_, _ = fmt.Fprintf(
					w,
					"        Started At:\t%s\n",
					status.State.Running.StartedAt.Format(time.RFC3339Nano),
				)
			}
		})
		if status.Resources != nil && len(status.Resources.Limits) > 0 {
			AppendTabbedString(w, func(w io.Writer) {
				_, _ = w.Write([]byte("      Limits:\n"))
				writeIfNotEmptyf(w, "        CPU:\t%s\n", status.Resources.Limits.Cpu().String())
				writeIfNotEmptyf(w, "        Memory:\t%s\n", status.Resources.Limits.Memory().String())
			})
		}
		if status.Resources != nil && len(status.Resources.Requests) > 0 {
			AppendTabbedString(w, func(w io.Writer) {
				_, _ = w.Write([]byte("      Requests:\n"))
				writeIfNotEmptyf(w, "        CPU:\t%s\n", status.Resources.Requests.Cpu().String())
				writeIfNotEmptyf(w, "        Memory:\t%s\n", status.Resources.Requests.Memory().String())
			})
		}
		if len(status.AllocatedResources) > 0 {
			AppendTabbedString(w, func(w io.Writer) {
				_, _ = w.Write([]byte("      Allocated:\n"))
				writeIfNotEmptyf(w, "        CPU:\t%s\n", status.AllocatedResources.Cpu())
				writeIfNotEmptyf(w, "        Memory:\t%s\n", status.AllocatedResources.Memory())
			})
		}
		if len(status.VolumeMounts) > 0 {
			_, _ = w.Write([]byte("      Volume Mounts:\n"))
			mounts := []string{}
			for _, mount := range status.VolumeMounts {
				perms := ""
				if mount.ReadOnly {
					perms = "(RO)"
				}
				mounts = append(mounts, fmt.Sprintf("        %s %s", mount.MountPath, perms))
			}
			slices.Sort(mounts)
			_, _ = w.Write([]byte(strings.Join(mounts, "\n") + "\n"))
		}
		if status.User != nil && status.User.Linux != nil {
			_, _ = w.Write([]byte("      Linux User:\n"))
			AppendTabbedString(w, func(w io.Writer) {
				_, _ = fmt.Fprintf(w, "        UID:\t%d\n", status.User.Linux.UID)
				_, _ = fmt.Fprintf(w, "        GID:\t%d\n", status.User.Linux.GID)
				if len(status.User.Linux.SupplementalGroups) > 0 {
					gids := make([]string, 0, len(status.User.Linux.SupplementalGroups))
					for _, gid := range status.User.Linux.SupplementalGroups {
						gids = append(gids, strconv.FormatInt(gid, 10))
					}
					_, _ = fmt.Fprintf(w, "        Additional GIDs:\t%s\n", strings.Join(gids, ", "))
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
		_, _ = fmt.Fprintf(w, "ID:\t%s\n", resource.ID)
		_, _ = fmt.Fprintf(w, "Name:\t%s\n", resource.Name)
		writeIfNotEmptyf(w, "Description:\t%s\n", resource.Description)
		_, _ = fmt.Fprintf(w, "Created By:\t%s\n", resource.CreatedBy)
		_, _ = fmt.Fprintf(
			w,
			"Created At:\t%s\n",
			TimePtrToStringWithAge(resource.CreatedAt),
		)
		_, _ = fmt.Fprintf(
			w,
			"Modified At:\t%s\n",
			TimePtrToStringWithAge(resource.ModifiedAt),
		)
		_, _ = fmt.Fprintf(w, "State:\t%s\n", resource.State)
		writeIfNotEmptyf(w, "Spec:\t%s\n", specAsString(resource.Spec))
		writeIfNotEmptyf(w, "Message:\t%s\n", resource.Message)
	})
}

// AppendEnvDetails appends the description of the given environment as a tabbed
// string to the given writer.
func AppendEnvDetails(
	w io.Writer,
	env map[string]string,
) {
	if len(env) > 0 {
		_, _ = w.Write([]byte("Environment:\n"))
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
		_, _ = fmt.Fprintf(w, "ID:\t%s\n", resource.ID)
		_, _ = fmt.Fprintf(w, "Controller ID:\t%s\n", resource.ControllerID)
		_, _ = fmt.Fprintf(w, "Name:\t%s\n", resource.Name)
		_, _ = fmt.Fprintf(w, "Created By:\t%s\n", resource.CreatedBy)
		_, _ = fmt.Fprintf(
			w,
			"Created At:\t%s\n",
			TimePtrToStringWithAge(resource.CreatedAt),
		)
		_, _ = fmt.Fprintf(w, "Image ID:\t%s\n", resource.ImageID)
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
		_, _ = fmt.Fprintf(w, "ID:\t%s\n", resource.ID)
		_, _ = fmt.Fprintf(w, "Controller ID:\t%s\n", resource.ControllerID)
		_, _ = fmt.Fprintf(w, "Name:\t%s\n", resource.Name)
		writeIfNotEmptyf(w, "Description:\t%s\n", resource.Description)
		_, _ = fmt.Fprintf(w, "Created By:\t%s (%s)\n", resource.Username, resource.Email)
		_, _ = fmt.Fprintf(
			w,
			"Created At:\t%s\n",
			TimePtrToStringWithAge(resource.CreatedAt),
		)
		_, _ = fmt.Fprintf(
			w,
			"Modified At:\t%s\n",
			TimePtrToStringWithAge(resource.ModifiedAt),
		)
		_, _ = fmt.Fprintf(w, "Image:\t%s\n", image)
		writeIfNotEmptyf(w, "Context:\t%s\n", resource.Context)
		writeIfNotEmptyf(w, "Spec:\t%s\n", specAsString(resource.Spec))
		writeIfNotEmptyf(w, "SubPath:\t%s\n", resource.SubPath)
		writeIfNotEmptyf(w, "Latest Image Build ID:\t%s\n", resource.LatestImageBuildID)
		writeIfNotEmptyf(w, "State\t%s\n", resource.CurrentState)
		writeIfNotEmptyf(w, "Message\t%s\n", resource.Message)
	})
}

// AppendJobDetails appends the description of the given types.Job as a tabbed
// string to the given writer.
func AppendJobDetails(
	w io.Writer,
	resource types.Job,
) {
	AppendTabbedString(w, func(w io.Writer) {
		_, _ = fmt.Fprintf(w, "ID:\t%s\n", resource.ID)
		_, _ = fmt.Fprintf(w, "Name:\t%s\n", resource.Name)
		writeIfNotEmptyf(w, "Description:\t%s\n", resource.Description)
		_, _ = fmt.Fprintf(w, "Created By:\t%s (%s)\n", resource.Username, resource.Email)
		_, _ = fmt.Fprintf(
			w,
			"Created At:\t%s\n",
			TimePtrToStringWithAge(resource.CreatedAt),
		)
		_, _ = fmt.Fprintf(
			w,
			"Modified At:\t%s\n",
			TimePtrToStringWithAge(resource.ModifiedAt),
		)
		if resource.MostRecentJobSubmissionAt != nil {
			writeIfNotEmptyf(
				w,
				"Most Recent Job Run At:\t%s\n",
				resource.MostRecentJobSubmissionAt.Format(time.RFC3339Nano),
			)
		}
		writeIfNotEmptyf(w, "Most Recent Job Run ID:\t%s\n", resource.MostRecentJobSubmissionID)
		_, _ = fmt.Fprintf(w, "Nodepool:\t%s\n", resource.Nodepool)
		writeIfNotEmptyf(w, "Application:\t%s\n", resource.ApplicationID)
		writeIfNotEmptyf(w, "Context:\t%s\n", resource.Context)
		writeIfNotEmptyf(w, "Spec:\t%s\n", specAsString(resource.Spec))
		writeIfNotEmptyf(w, "SubPath:\t%s\n", resource.SubPath)
		writeIfNotEmptyf(w, "Schedule:\t%s\n", resource.Schedule)
	})
}

// AppendJobSubmissionDetails appends the description of the given
// types.JobSubmission as a tabbed string to the given writer.
func AppendJobSubmissionDetails(
	w io.Writer,
	resource types.JobSubmission,
) {
	AppendTabbedString(w, func(w io.Writer) {
		_, _ = fmt.Fprintf(w, "ID:\t%s\n", resource.ID)
		_, _ = fmt.Fprintf(w, "Controller ID:\t%s\n", resource.ControllerID)
		_, _ = fmt.Fprintf(w, "Job Name:\t%s\n", resource.JobName)
		_, _ = fmt.Fprintf(w, "Job ID:\t%s\n", resource.JobID)
		_, _ = fmt.Fprintf(w, "URL:\t%s\n", resource.URL)
		_, _ = fmt.Fprintf(w, "Log URL:\t%s\n", resource.LogURL)
		_, _ = fmt.Fprintf(w, "Created By:\t%s (%s)\n", resource.Username, resource.Email)
		_, _ = fmt.Fprintf(
			w,
			"Created At:\t%s\n",
			TimePtrToStringWithAge(resource.CreatedAt),
		)
		_, _ = fmt.Fprintf(
			w,
			"Modified At:\t%s\n",
			TimePtrToStringWithAge(resource.ModifiedAt),
		)
		if resource.CanceledAt != nil {
			_, _ = fmt.Fprintf(
				w,
				"Canceled At:\t%s\n",
				TimePtrToStringWithAge(resource.CanceledAt),
			)
		}
		if resource.CompletedAt != nil {
			_, _ = fmt.Fprintf(
				w,
				"Completed At:\t%s\n",
				TimePtrToStringWithAge(resource.CompletedAt),
			)
		}
		_, _ = fmt.Fprintf(w, "Runtime:\t%f\n", resource.Runtime)
		writeIfNotEmptyf(w, "Nodepool:\t%s\n", resource.Nodepool)
		_, _ = fmt.Fprintf(w, "Scheduled:\t%t\n", resource.Scheduled)
		_, _ = fmt.Fprintf(w, "State:\t%s\n", resource.CurrentState)
		writeIfNotEmptyf(w, "Message:\t%s\n", resource.Message)
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
		_, _ = fmt.Fprintf(w, "Labels:\n")
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
				writeIfNotEmptyf(w, "Pod %s:\n", ps.Name)
				writeIfNotEmptyf(w, "  Phase:\t%s\n", string(ps.Status.Phase))
				if ps.Deleted {
					_, _ = w.Write([]byte("  Deleted:\ttrue\n"))
				}
				writeIfNotEmptyf(w, "  Message:\t%s\n", ps.Status.Message)
				writeIfNotEmptyf(w, "  Reason:\t%s\n", ps.Status.Reason)
				writeIfNotEmptyf(w, "  Nominated node:\t%s\n", ps.Status.NominatedNodeName)
				writeIfNotEmptyf(w, "  IPs:\t%s\n", formatIPs(ps.Status.PodIP, ps.Status.PodIPs))
			})
			AppendTabbedString(w, func(w io.Writer) {
				if len(ps.Status.Conditions) > 0 {
					_, _ = w.Write([]byte("  Conditions:\n"))
					for _, cond := range ps.Status.Conditions {
						_, _ = fmt.Fprintf(w, "    %s\t%s\n", cond.Type, cond.Status)
					}
				}
			})
			AppendContainerStatuses(w, "Init Containers:", ps.Status.InitContainerStatuses)
			AppendContainerStatuses(w, "Containers:", ps.Status.ContainerStatuses)
			AppendContainerStatuses(w, "Ephemeral Containers:", ps.Status.EphemeralContainerStatuses)
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
	_, _ = w.Write([]byte(constants.DescribeLabelProperties + ":\n"))
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
		_, _ = w.Write([]byte(constants.DescribeLabelRoles + ":\n"))
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

// AppendServiceRequestDetails appends the description of the given
// types.ServiceRequest slice as a tabbed string to the given writer.
func AppendServiceRequestDetails(
	w io.Writer,
	header string,
	services []types.ServiceRequest,
) {
	if len(services) > 0 {
		_, _ = w.Write([]byte(header + ":\n"))
		for i, svc := range services {
			AppendTabbedString(w, func(w io.Writer) {
				_, _ = fmt.Fprintf(w, "  %d:\n", i)
				if svc.ID != "" {
					_, _ = fmt.Fprintf(w, "    ID:\t%s\n", svc.ID)
				}
				if svc.Image != "" {
					_, _ = fmt.Fprintf(w, "    Image:\t%s\n", svc.Image)
				}
				if len(svc.Env) > 0 {
					_, _ = fmt.Fprint(w, "    Environment:\n")
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
		_, _ = w.Write([]byte(constants.DescribeLabelMetadata + ":\n"))
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
		_, _ = fmt.Fprintf(w, "ID:\t%s\n", resource.ID)
		_, _ = fmt.Fprintf(w, "Controller ID:\t%s\n", resource.ControllerID)
		_, _ = fmt.Fprintf(w, "Name:\t%s\n", resource.Name)
		_, _ = fmt.Fprintf(w, "Description:\t%s\n", resource.Description)
		_, _ = fmt.Fprintf(w, "Created By:\t%s (%s)\n", resource.Username, resource.Email)
		_, _ = fmt.Fprintf(w,
			"Created At:\t%s\n",
			TimePtrToStringWithAge(resource.CreatedAt),
		)
		_, _ = fmt.Fprintf(
			w,
			"Modified At:\t%s\n",
			TimePtrToStringWithAge(resource.ModifiedAt),
		)
		writeIfNotEmptyf(w, "Spec:\t%s\n", snapshotSpecAsString(resource.Spec))
		writeIfNotEmptyf(w, "Message:\t%s\n", resource.Message)
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
		writeIfNotEmptyf(w, prefix+"Access Modes:\t%s\n", strings.Join(accessModes, ", "))
		writeIfNotEmptyf(w, prefix+"Storage Class:\t%s\n", pointer.UnPtr(spec.StorageClassName))
		writeIfNotEmptyf(w, prefix+"Size:\t%s\n", spec.Resources.Requests.Storage().String())
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
		_, _ = fmt.Fprintf(w, "ID:\t%s\n", user.ID)
		_, _ = fmt.Fprintf(w, "Database ID:\t%s\n", user.DatabaseID)
		_, _ = fmt.Fprintf(
			w,
			"Created At:\t%s\n",
			TimePtrToStringWithAge(&user.CreationTimeStamp),
		)
		_, _ = fmt.Fprintf(w, "Username:\t%s\n", user.Name)
		_, _ = fmt.Fprintf(w, "Email:\t%s\n", user.Email)
	})
}

// AppendVolumeDetails appends the description of the given types.VolumeStatus
// as a tabbed string to the given writer.
func AppendVolumeDetails(
	w io.Writer,
	volumeStatus *types.VolumeStatus,
) {
	_, _ = w.Write([]byte("Volume Status:\n"))
	AppendTabbedString(w, func(w io.Writer) {
		if volumeStatus != nil {
			_, _ = fmt.Fprintf(w, "  Ready To Use:\t%v\n", pointer.UnPtr(volumeStatus.Status.ReadyToUse))
			_, _ = fmt.Fprintf(
				w,
				"  Restore Size:\t%v\n",
				QuantityPtrToString(volumeStatus.Status.RestoreSize),
			)
			_, _ = fmt.Fprintf(
				w,
				"  Content Name:\t%v\n",
				pointer.UnPtr(volumeStatus.Status.BoundVolumeSnapshotContentName),
			)
			_, _ = fmt.Fprintf(
				w,
				"  Creation Time:\t%v\n",
				metaTimePtrToString(volumeStatus.Status.CreationTime),
			)
		} else {
			_, _ = w.Write([]byte("  Ready To Use:\tfalse\n"))
			_, _ = w.Write([]byte("  Restore Size:\t<unknown>\n"))
			_, _ = w.Write([]byte("  Content Name:\t<unknown>\n"))
			_, _ = w.Write([]byte("  Creation Time:\t<unknown>\n"))
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
		_, _ = fmt.Fprintf(w, "ID:\t%s\n", resource.ID)
		_, _ = fmt.Fprintf(w, "Controller ID:\t%s\n", resource.ControllerID)
		_, _ = fmt.Fprintf(w, "Name:\t%s\n", resource.Name)
		_, _ = fmt.Fprintf(w, "URL:\t%s\n", resource.URL)
		_, _ = fmt.Fprintf(w, "Created By:\t%s (%s)\n", resource.Username, resource.Email)
		_, _ = fmt.Fprintf(
			w,
			"Created At:\t%s\n",
			TimePtrToStringWithAge(resource.CreatedAt),
		)
		_, _ = fmt.Fprintf(
			w,
			"Modified At:\t%s\n",
			TimePtrToStringWithAge(resource.ModifiedAt),
		)
		_, _ = fmt.Fprintf(w, "Nodepool:\t%s\n", resource.Nodepool)
		_, _ = fmt.Fprintf(w, "Desired State:\t%s\n", resource.DesiredState)
		_, _ = fmt.Fprintf(w, "Current State:\t%s\n", resource.CurrentState)
		writeIfNotEmptyf(w, "Message:\t%s\n", resource.Message)
		writeIfNotEmptyf(w, "Application:\t%s\n", resource.ApplicationID)
	})
}

// AppendWorkspaceDetails appends the description of the given types.Workspace
// as a tabbed string to the given writer.
func AppendWorkspaceDetails(
	w io.Writer,
	resource types.Workspace,
) {
	AppendTabbedString(w, func(w io.Writer) {
		_, _ = fmt.Fprintf(w, "ID:\t%s\n", resource.ID)
		_, _ = fmt.Fprintf(w, "Controller ID:\t%s\n", resource.ControllerID)
		_, _ = fmt.Fprintf(w, "Name:\t%s\n", resource.Name)
		writeIfNotEmptyf(w, "Description:\t%s\n", resource.Description)
		_, _ = fmt.Fprintf(w, "Storage Type:\t%s\n", resource.StorageType)
		_, _ = fmt.Fprintf(w, "Created By:\t%s (%s)\n", resource.Username, resource.Email)
		_, _ = fmt.Fprintf(
			w,
			"Created At:\t%s\n",
			TimePtrToStringWithAge(resource.CreatedAt),
		)
		_, _ = fmt.Fprintf(
			w,
			"Modified At:\t%s\n",
			TimePtrToStringWithAge(resource.ModifiedAt),
		)
		_, _ = fmt.Fprintf(w, "State:\t%s\n", resource.State)
		writeIfNotEmptyf(w, "Message:\t%s\n", resource.Message)
		writeIfNotEmptyf(w, "Phase:\t%s\n", resource.PVCStatus.Status.Phase)
	})
}

// AppendWorkspaceRequestsDetails appends the description of the given
// types.WorkspaceRequest slice as a tabbed string to the given writer.
func AppendWorkspaceRequestsDetails(
	w io.Writer,
	workspaces []types.WorkspaceRequest,
) {
	if len(workspaces) > 0 {
		_, _ = w.Write([]byte("Volumes:\n"))
		for i, wksp := range workspaces {
			AppendTabbedString(w, func(w io.Writer) {
				_, _ = fmt.Fprintf(w, "  %d:\n", i)
				if wksp.ID != "" {
					_, _ = fmt.Fprintf(w, "    ID:\t%s\n", wksp.ID)
				}
				if wksp.MountName != "" {
					_, _ = fmt.Fprintf(w, "    Mount name:\t%s\n", wksp.MountName)
				}
				if wksp.EphemeralPVC != nil {
					_, _ = w.Write([]byte("    Ephemeral PVC:\n"))
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
		return constants.ValueUnknown
	}
	return fmt.Sprintf("%s (%s ago)", t.UTC().Format(time.RFC3339), timeutil.Since(t))
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
		return "Waiting"
	case state.Terminated != nil:
		return "Terminated"
	case state.Running != nil:
		return "Running"
	default:
		return constants.ValueUnknown
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
		return constants.ValueUnknown
	}
	return TimePtrToStringWithAge(&t.Time)
}

func snapshotSpecAsString(v *snapshot1.VolumeSnapshotSpec) string {
	if v == nil {
		return ""
	}
	r, err := marshalIndented(v, "                 ")
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return r
}

func specAsString(v map[string]interface{}) string {
	if len(v) == 0 {
		return ""
	}
	r, err := marshalIndented(v, "               ")
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return strings.TrimSpace(r)
}

// writeIfNotEmptyf writes to the given io.Writer if the given value is not
// empty.
func writeIfNotEmptyf(
	w io.Writer,
	format string,
	value ...any,
) {
	if len(value) != 0 && value[0] != "" {
		_, _ = fmt.Fprintf(w, format, value...)
	}
}
