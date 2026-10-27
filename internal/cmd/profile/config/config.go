// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/sassoftware/sas-ggdk/pkg/errors"
	"github.com/sassoftware/wbctl/internal/configutil"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/logutil"
	"github.com/sassoftware/wbctl/internal/wbctljson"
	"github.com/spf13/afero"
	"go.yaml.in/yaml/v3"
)

// ReadDocument reads the document at the given path into a generic map. A
// missing file is treated as an empty document so callers can bootstrap a new
// config file.
func ReadDocument(
	fsys afero.Fs,
	configPath string,
) (map[string]any, error) {
	logutil.GetLogger().Info("reading config", "path", configPath)
	exists, err := afero.Exists(fsys, configPath)
	if err != nil {
		return nil, errors.Wrap(err, i18n.Format(i18n.ConfigFileReadFailed, i18n.A{constants.I18nArgKeyPath: configPath}))
	}
	if !exists {
		return map[string]any{}, nil
	}
	content, err := afero.ReadFile(fsys, configPath)
	if err != nil {
		return nil, errors.Wrap(err, i18n.Format(i18n.ConfigFileReadFailed, i18n.A{constants.I18nArgKeyPath: configPath}))
	}
	doc := map[string]any{}
	format, err := detectWritableFormat(configPath)
	if err != nil {
		return nil, errors.Wrap(err, i18n.Format(i18n.ConfigFileReadFailed, i18n.A{constants.I18nArgKeyPath: configPath}))
	}
	if format == formatYAML {
		err = yaml.Unmarshal(content, &doc)
	} else {
		err = json.Unmarshal(content, &doc)
	}
	if err != nil {
		return nil, errors.Wrap(err, i18n.Format(i18n.ConfigFileReadFailed, i18n.A{constants.I18nArgKeyPath: configPath}))
	}
	return doc, nil
}

// WriteDocument marshals doc as the given format and writes it to configPath,
// creating the parent directory if needed. The directory and file are written
// with restrictive permissions since the document may contain credentials.
func WriteDocument(
	fsys afero.Fs,
	configPath string,
	doc map[string]any,
) error {
	logutil.GetLogger().Info("writing config", "path", configPath)
	dir := filepath.Dir(configPath)
	if dir != "" && dir != "." {
		err := fsys.MkdirAll(dir, os.FileMode(constants.ConfigFileDirPerm))
		if err != nil {
			return errors.Wrap(err, i18n.Format(i18n.ConfigFileWriteFailed, i18n.A{constants.I18nArgKeyPath: configPath}))
		}
	}
	var content []byte
	var err error
	format, err := detectWritableFormat(configPath)
	if err != nil {
		return errors.Wrap(err, i18n.Format(i18n.ConfigFileWriteFailed, i18n.A{constants.I18nArgKeyPath: configPath}))
	}
	if format == formatYAML {
		content, err = yaml.Marshal(doc)
	} else {
		content, err = wbctljson.MakeMarshaler[map[string]any](wbctljson.ViewOptions)(doc)
	}
	if err != nil {
		return errors.Wrap(err, i18n.Format(i18n.ConfigFileWriteFailed, i18n.A{constants.I18nArgKeyPath: configPath}))
	}
	filePerm := os.FileMode(constants.ConfigFilePerm)
	err = afero.WriteFile(fsys, configPath, content, filePerm)
	if err != nil {
		return errors.Wrap(err, i18n.Format(i18n.ConfigFileWriteFailed, i18n.A{constants.I18nArgKeyPath: configPath}))
	}
	// afero.WriteFile only applies perm at file-creation time, so explicitly
	// tighten permissions in case the file already existed with looser perms.
	// The Afero memmapped file system implementation of Chmod does not clean
	// the path so that must be done here for tests.
	err = fsys.Chmod(filepath.Clean(configPath), filePerm)
	if err != nil {
		return errors.Wrap(err, i18n.Format(i18n.ConfigFileWriteFailed, i18n.A{constants.I18nArgKeyPath: configPath}))
	}
	return nil
}

type format int

const (
	formatYAML format = iota
	formatJSON
)

// detectWritableFormat determines the Format to use when writing the given
// path. Files with a .yaml/.yml extension are written as YAML. Any other
// extension that viper recognizes but that this package cannot safely write
// (for example .toml or .ini) results in an error. All other extensions,
// including no extension at all, default to JSON.
func detectWritableFormat(configPath string) (format, error) {
	ext := filepath.Ext(configPath)
	if ext != "" {
		ext = ext[1:]
	}
	switch ext {
	case constants.ConfigFileTypeYAML, constants.ConfigFileTypeYML:
		return formatYAML, nil
	case constants.ConfigFileTypeDefault:
		return formatJSON, nil
	}
	if configutil.IsKnownFormat(configPath) {
		return formatJSON, errors.New(
			i18n.Format(i18n.ConfigFileUnsupportedFormat, i18n.A{constants.I18nArgKeyPath: configPath}),
		)
	}
	return formatJSON, nil
}
