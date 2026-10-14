// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"maps"
	"slices"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/ui"
)

// API is the OpenAPI (Swagger) specification for the Workbench service.
type API *spec.Swagger

// GetAPI returns the API from the WorkbenchService.
func GetAPI(
	ctx context.Context,
	s *WorkbenchService,
) (API, error) {
	return result.Split(Get[API](ctx, s, postgrest.Endpoint(constants.WorkbenchEndpointRoot), nil))
}

// PrintAPI prints the given API.
func PrintAPI(
	_ui ui.UI,
	resources API,
) error {
	if resources.Paths == nil {
		return nil
	}
	pathData := map[string]spec.PathItem{}
	for path, data := range resources.Paths.Paths {
		pathData[path] = data
	}
	sortedPaths := slices.Collect(maps.Keys(pathData))
	slices.Sort(sortedPaths)
	tbl := _ui.Table([]string{"PATH", "VERB", constants.TableLabelDescription})
	for _, path := range sortedPaths {
		data := pathData[path]
		verbs := getVerbs(data)
		description := getDescription(data)
		tbl.Add(
			path,
			verbs,
			description,
		)
	}
	tbl.Print()
	return nil
}

func getVerbs(desc spec.PathItem) string {
	verbs := []string{}
	if desc.Get != nil {
		verbs = append(verbs, "GET")
	}
	if desc.Post != nil {
		verbs = append(verbs, "POST")
	}
	return strings.Join(verbs, ", ")
}

func getDescription(desc spec.PathItem) string {
	description := []string{}
	if desc.Get != nil {
		d := strings.TrimSpace(desc.Get.Summary)
		if d != "" {
			description = append(description, d)
		}
	}
	if desc.Post != nil {
		d := strings.TrimSpace(desc.Post.Summary)
		if d != "" {
			description = append(description, d)
		}
	}
	return strings.Join(description, ", ")
}
