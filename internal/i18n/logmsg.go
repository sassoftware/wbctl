// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package i18n

import (
	"os"
	"sync"

	"github.com/gotnospirit/messageformat"
	"github.com/magiconair/properties"
	"github.com/sassoftware/wbctl/internal/exiter"
	"github.com/sassoftware/wbctl/internal/ui"
)

// CLI messages.
const (
	BundleID                             = "wbctl-cli-icu"
	APIRequestBodyType                   = BundleID + "api.request.body.type.msg"
	AtLeastOneResourceIDForCommand       = BundleID + ".at.least.one.resource.id.for.command.msg"
	Auth0OnlyOneSortFieldAllowed         = BundleID + ".auth.only.one.sort.field.allowed.msg"
	AuthModeDeviceRequiresOrganization   = BundleID + ".auth.mode.device.requires.organization.msg"
	BulkUpdateNameError                  = BundleID + ".bulk.update.name.error.msg"
	CachedAccessTokenIsExpired           = BundleID + ".cached.access.token.is.expired.msg"
	CmdAPIUsage                          = BundleID + ".cmd.api.usage.msg"
	CmdCancelDescription                 = BundleID + ".cmd.cancel.desc"
	CmdCancelTypeDescription             = BundleID + ".cmd.cancel.type.description.msg"
	CmdCancelTypeUsage                   = BundleID + ".cmd.cancel.type.usage.msg"
	CmdCancelUsage                       = BundleID + ".cmd.cancel.usage.msg"
	CmdCreateDescription                 = BundleID + ".cmd.create.desc"
	CmdCreateTypeDescription             = BundleID + ".cmd.create.type.description.msg"
	CmdCreateTypeUsage                   = BundleID + ".cmd.create.type.usage.msg"
	CmdCreateUsage                       = BundleID + ".cmd.create.usage.msg"
	CmdDeleteDescription                 = BundleID + ".cmd.delete.desc"
	CmdDeleteTypeDescription             = BundleID + ".cmd.delete.type.description.msg"
	CmdDeleteTypeUsage                   = BundleID + ".cmd.delete.type.usage.msg"
	CmdDeleteUsage                       = BundleID + ".cmd.delete.usage.msg"
	CmdDescribeDescription               = BundleID + ".cmd.describe.desc"
	CmdDescribeTypeDescription           = BundleID + ".cmd.describe.type.description.msg"
	CmdDescribeTypeUsage                 = BundleID + ".cmd.describe.type.usage.msg"
	CmdDescribeUsage                     = BundleID + ".cmd.describe.usage.msg"
	CmdEditDescription                   = BundleID + ".cmd.edit.description.msg"
	CmdEditTypeDescription               = BundleID + ".cmd.edit.type.description.msg"
	CmdEditTypeUsage                     = BundleID + ".cmd.edit.type.usage.msg"
	CmdEditUsage                         = BundleID + ".cmd.edit.usage.msg"
	CmdGetDescription                    = BundleID + ".cmd.get.desc"
	CmdGetTypeDescription                = BundleID + ".cmd.get.type.description.msg"
	CmdGetTypeUsage                      = BundleID + ".cmd.get.type.usage.msg"
	CmdGetUsage                          = BundleID + ".cmd.get.usage.msg"
	CmdLogDescription                    = BundleID + ".cmd.log.desc"
	CmdLogTypeDescription                = BundleID + ".cmd.log.type.description.msg"
	CmdLogTypeUsage                      = BundleID + ".cmd.log.type.usage.msg"
	CmdLogUsage                          = BundleID + ".cmd.log.usage.msg"
	CmdLoginDescription                  = BundleID + ".cmd.login.desc"
	CmdLoginUsage                        = BundleID + ".cmd.login.usage.msg"
	CmdRebuildDescription                = BundleID + ".cmd.rebuild.desc"
	CmdRebuildTypeDescription            = BundleID + ".cmd.rebuild.type.description.msg"
	CmdRebuildTypeUsage                  = BundleID + ".cmd.rebuild.type.usage.msg"
	CmdRebuildUsage                      = BundleID + ".cmd.rebuild.usage.msg"
	CmdSettingsDescription               = BundleID + ".cmd.settings.description.msg"
	CmdSettingsUsage                     = BundleID + ".cmd.settings.usage.msg"
	CmdStartDescription                  = BundleID + ".cmd.start.desc"
	CmdStartTypeDescription              = BundleID + ".cmd.start.type.description.msg"
	CmdStartTypeUsage                    = BundleID + ".cmd.start.type.usage.msg"
	CmdStartUsage                        = BundleID + ".cmd.start.usage.msg"
	CmdStopDescription                   = BundleID + ".cmd.stop.desc"
	CmdStopTypeDescription               = BundleID + ".cmd.stop.type.description.msg"
	CmdStopTypeUsage                     = BundleID + ".cmd.stop.type.usage.msg"
	CmdStopUsage                         = BundleID + ".cmd.stop.usage.msg"
	CmdUpdateDescription                 = BundleID + ".cmd.update.desc"
	CmdUpdateTypeDescription             = BundleID + ".cmd.update.type.description.msg"
	CmdUpdateTypeUsage                   = BundleID + ".cmd.update.type.usage.msg"
	CmdUpdateUsage                       = BundleID + ".cmd.update.usage.msg"
	ConfigFileReadFailed                 = BundleID + ".config.file.read.failed.msg"
	DataplaneNotFound                    = BundleID + ".dataplane.not.found.msg"
	DeviceAuthorizationTimeout           = BundleID + ".device.authorization.timeout.msg"
	DeviceTokenPollCancelled             = BundleID + ".device.token.poll.cancelled.msg"
	EditCanceledFileEmpty                = BundleID + ".edit.canceled.file.empty.msg"
	EditCanceledFileUnchanged            = BundleID + ".edit.canceled.file.unchanged.msg"
	EitherApplicationOrServiceIsRequired = BundleID + ".either.application.or.service.is.required.msg"
	EitherFlagNamesAreRequired           = BundleID + ".either.flag.names.are.required.msg"
	EndpointMustBeSpecified              = BundleID + ".endpoint.must.be.specified.msg"
	Error                                = BundleID + ".error.msg"
	FailedToAddSnapshot                  = BundleID + ".failed.to.add.snapshot.msg"
	FailedToGetCSRFToken                 = BundleID + ".failed.to.get.csrf.token.msg"
	FailedToGetDeviceToken               = BundleID + ".failed.to.get.device.token.msg"
	FailedToGetPKCEToken                 = BundleID + ".failed.to.get.pkce.token.msg"
	FailedToGetWorkspace                 = BundleID + ".failed.to.get.workspace.msg"
	FailedToLoadTokenCache               = BundleID + ".failed.to.load.token.cache.msg"
	FailedToRefreshToken                 = BundleID + ".failed.to.refresh.token.msg"
	FailedToSaveDeviceToken              = BundleID + ".failed.to.save.device.token.msg"
	FailedToSaveTokenCache               = BundleID + ".failed.to.save.token.cache.msg"
	FailedToStartEditor                  = BundleID + ".failed.to.start.editor.msg"
	FailedToWaitForEditor                = BundleID + ".failed.to.wait.for.editor.msg"
	FallingBackToFullReauthorization     = BundleID + ".falling.back.to.full.reauthorization.msg"
	FlagApplicationUsage                 = BundleID + ".flag.application.usage.msg"
	FlagAuthModeUsage                    = BundleID + ".flag.auth.mode.usage.msg"
	FlagClientIDUsage                    = BundleID + ".flag.client.id.usage.msg"
	FlagCommandUsage                     = BundleID + ".flag.command.usage.msg"
	FlagContextUsage                     = BundleID + ".flag.context.usage.msg"
	FlagDataUsage                        = BundleID + ".flag.data.usage.msg"
	FlagDataplaneUsage                   = BundleID + ".flag.dataplane.usage.msg"
	FlagDescriptionUsage                 = BundleID + ".flag.description.usage.msg"
	FlagDesiredStateUsage                = BundleID + ".flag.desired.state.usage.msg"
	FlagDryRunUsage                      = BundleID + ".flag.dry.run.usage.msg"
	FlagEndpointUsage                    = BundleID + ".flag.endpoint.usage.msg"
	FlagEnvironmentUsage                 = BundleID + ".flag.environment.usage.msg"
	FlagFilenameUsage                    = BundleID + ".flag.filename.usage.msg"
	FlagHomeDirUsage                     = BundleID + ".flag.home.dir.usage.msg"
	FlagIgnoreTokenCacheFlagUsage        = BundleID + ".flag.ignore.token.cache.flag.usage.msg"
	FlagImageUsage                       = BundleID + ".flag.image.usage.msg"
	FlagIngressHostUsage                 = BundleID + ".flag.ingress.host.usage.msg"
	FlagInitializerUsage                 = BundleID + ".flag.initializer.usage.msg"
	FlagIsRequired                       = BundleID + ".flag.is.required.msg"
	FlagJobIDUsage                       = BundleID + ".flag.job.id.usage.msg"
	FlagLabelUsage                       = BundleID + ".flag.label.usage.msg"
	FlagMethodUsage                      = BundleID + ".flag.method.usage.msg"
	FlagNameUsage                        = BundleID + ".flag.name.usage.msg"
	FlagNoListenUsage                    = BundleID + ".flag.no.listen.usage.msg"
	FlagNodepoolIsRequired               = BundleID + ".flag.nodepool.is.required.msg"
	FlagNodepoolUsage                    = BundleID + ".flag.nodepool.usage.msg"
	FlagOrganizationUsage                = BundleID + ".flag.organization.usage.msg"
	FlagOutputUsage                      = BundleID + ".flag.output.usage.msg"
	FlagQueryUsage                       = BundleID + ".flag.query.usage.msg"
	FlagRequiredWhen                     = BundleID + ".flag.required.when.msg"
	FlagScheduleUsage                    = BundleID + ".flag.schedule.usage.msg"
	FlagServiceUsage                     = BundleID + ".flag.service.usage.msg"
	FlagSetUsage                         = BundleID + ".flag.set.usage.msg"
	FlagSizeUsage                        = BundleID + ".flag.size.usage.msg"
	FlagSnapshotUsage                    = BundleID + ".flag.snapshot.usage.msg"
	FlagSnapshotVolumeUsage              = BundleID + ".flag.snapshot.volume.usage.msg"
	FlagSnapshotWorkspaceUsage           = BundleID + ".flag.snapshot.workspace.usage.msg"
	FlagSortUsage                        = BundleID + ".flag.sort.usage.msg"
	FlagSpecUsage                        = BundleID + ".flag.spec.usage.msg"
	FlagStorageClassUsage                = BundleID + ".flag.storage.class.usage.msg"
	FlagSubPathUsage                     = BundleID + ".flag.sub.path.usage.msg"
	FlagTenantUsage                      = BundleID + ".flag.tenant.usage.msg"
	FlagUserIDUsage                      = BundleID + ".flag.user.id.usage.msg"
	FlagValueUsage                       = BundleID + ".flag.value.usage.msg"
	FlagVolumeUsage                      = BundleID + ".flag.volume.usage.msg"
	FlagWorkspaceDeprecationUsage        = BundleID + ".flag.workspace.deprecation.usage.msg"
	FlagWorkspaceUsage                   = BundleID + ".flag.workspace.usage.msg"
	IDRequired                           = BundleID + ".id.required.msg"
	InvalidAuthMode                      = BundleID + ".invalid.auth.mode.msg"
	InvalidSnapshotAccessModeMissing     = BundleID + ".invalid.snapshot.access.mode.missing.msg"
	LoginAuthModePrompt                  = BundleID + ".login.authmode.prompt.msg"
	LoginEndpointPrompt                  = BundleID + ".login.endpoint.prompt.msg"
	MultipleDataplanesFoundSpecifyOne    = BundleID + ".multiple.dataplanes.found.specify.one.msg"
	MultipleElementsFound                = BundleID + ".multiple.elements.found.msg"
	NoAccessTokenInCache                 = BundleID + ".no.access.token.in.cache.msg"
	NoElementsFound                      = BundleID + ".no.elements.found.msg"
	OneOfFlagNamesAreRequired            = BundleID + ".one.of.flag.names.are.required.msg"
	OnlyOneIDSupported                   = BundleID + ".only.one.id.supported.msg"
	OpenThisURLInYourBrowser             = BundleID + ".open.this.url.in.your.browser.msg"
	OpenURLAndEnterCode                  = BundleID + ".open.url.and.enter.code.msg"
	PKCEAuthorizationTimeout             = BundleID + ".pkce.authorization.timeout.msg"
	PKCEBrowserURLPrompt                 = BundleID + ".pkce.browser.url.prompt.msg"
	PKCEErrorResponse                    = BundleID + ".pkce.error.response.msg"
	PKCEInstructions                     = BundleID + ".pkce.instructions.msg"
	PKCEInstructionsWarning              = BundleID + ".pkce.instructions.warning.msg"
	PKCEReceivedStateMismatch            = BundleID + ".pkce.received.state.mismatch.msg"
	ProvidedRefreshTokenIsEmpty          = BundleID + ".provided.refresh.token.is.empty.msg"
	ResourceNotFoundWithID               = BundleID + ".resource.not.found.with.id.msg"
	ResourceTypeIsRequired               = BundleID + ".resource.type.is.required.msg"
	SnapshotNotReadyRestoreSizeMissing   = BundleID + ".snapshot.not.ready.restore.size.missing.msg"
	SnapshotNotReadySourceStorageMissing = BundleID + ".snapshot.not.ready.source.storage.missing.msg"
	SnapshotNotReadyVolumeStatusMissing  = BundleID + ".snapshot.not.ready.volume.status.missing.msg"
	StorageClassNotFound                 = BundleID + ".storage.class.not.found.msg"
	SupportedTypes                       = BundleID + ".supported.types.msg"
	TokenResponseMissingAccessToken      = BundleID + ".token.response.missing.access.token.msg"
	UnknownResourceType                  = BundleID + ".unknown.resource.type.msg"
	UnsupportedMapKeyKind                = BundleID + ".unsupported.map.key.kind.msg"
	UnsupportedMethod                    = BundleID + ".unsupported.method.msg"
	UnsupportedSortKeys                  = BundleID + ".unsupported.sort.keys.msg"
	UnsupportedType                      = BundleID + ".unsupported.type.msg"
	WarningDeafultStorageClass           = BundleID + ".warning.deafult.storage.class.msg"
	WarningNoVolumeSpecified             = BundleID + ".warning.no.volume.specified.msg"
)

var bundle *properties.Properties
var parser *messageformat.Parser
var once = sync.Once{}

// Bundle returns the workbench ops i18n message bundle.
func Bundle() *properties.Properties {
	once.Do(initialize)
	return bundle
}

// A defines a map of arguments to the Format function. It exists so that
// callers do not need to specify args with the full map[string]interface{}
// signature.
type A map[string]interface{}

// Format is a helper function that returns a translated string.
func Format(
	id string,
	a ...A,
) string {
	once.Do(initialize)
	var args map[string]interface{}
	if len(a) == 1 && a[0] != nil {
		args = a[0]
	}
	format, err := parser.Parse(bundle.GetString(id, id))
	if err != nil {
		return err.Error()
	}
	msg, err := format.FormatMap(args)
	if err != nil {
		return err.Error()
	}
	return msg
}

func initialize() {
	b, err := properties.Load([]byte(wbctlMessages), properties.UTF8)
	if err != nil {
		_ui := ui.NewUI(os.Stdin, os.Stdout, os.Stderr)
		_ui.Error("failed to load message properties: %w", err)
		exiter.OsExiter(1)
	}
	p, err := messageformat.NewWithCulture(culture)
	if err != nil {
		_ui := ui.NewUI(os.Stdin, os.Stdout, os.Stderr)
		_ui.Error("failed to load message parser: %w", err)
		exiter.OsExiter(1)
	}
	bundle = b
	parser = p
}

// Test hooks

var culture = "en"
