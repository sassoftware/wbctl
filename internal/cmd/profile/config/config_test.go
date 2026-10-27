// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sassoftware/sas-ggdk/pkg/errors"
	"github.com/sassoftware/wbctl/internal/cmd/profile/config"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestReadWriteDocument(t *testing.T) {
	t.Run("json round trip preserves other profiles and tightens permissions", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		path := filepath.Clean("/home/user/.wbctl/config.json")
		require.NoError(t, fsys.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, afero.WriteFile(
			fsys,
			path,
			[]byte(`{"verbose": "trace", "other":{"endpoint":"https://other.example.com"}}`),
			0o644,
		))

		doc, err := config.ReadDocument(fsys, path)
		require.NoError(t, err)
		doc["mine"] = map[string]any{"endpoint": "https://example.com"}
		require.NoError(t, config.WriteDocument(fsys, path, doc))

		doc2, err := config.ReadDocument(fsys, path)
		require.NoError(t, err)
		require.Equal(t, map[string]any{"endpoint": "https://other.example.com"}, doc2["other"])
		require.Equal(t, map[string]any{"endpoint": "https://example.com"}, doc2["mine"])
		require.Equal(t, "trace", doc2["verbose"])

		info, err := fsys.Stat(path)
		require.NoError(t, err)
		require.Equal(t, os.FileMode(0o600), info.Mode())
	})

	t.Run("yaml round trip", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		path := filepath.Clean("/home/user/.wbctl/config.yaml")
		doc := map[string]any{"mine": map[string]any{"endpoint": "https://example.com"}}
		require.NoError(t, config.WriteDocument(fsys, path, doc))

		doc2, err := config.ReadDocument(fsys, path)
		require.NoError(t, err)
		require.Equal(t, map[string]any{"endpoint": "https://example.com"}, doc2["mine"])
	})
}

func TestReadDocument(t *testing.T) {
	t.Run("missing file bootstraps empty document", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		doc, err := config.ReadDocument(fsys, "/home/user/.wbctl/config.json")
		require.NoError(t, err)
		require.Empty(t, doc)
	})

	t.Run("exists failure is returned", func(t *testing.T) {
		fsys := stubs.StubFS{
			Fs:       afero.NewMemMapFs(),
			StatFunc: func(afero.Fs, string) (os.FileInfo, error) { return nil, errors.New("stat failed") },
		}
		doc, err := config.ReadDocument(fsys, "/home/user/.wbctl/config.json")
		require.EqualError(t, err, "Failed to read config file '/home/user/.wbctl/config.json'\n"+
			" Caused by:\n"+
			"\t* stat failed")
		require.Empty(t, doc)
	})

	t.Run("read failure is returned", func(t *testing.T) {
		fsys := stubs.StubFS{
			Fs:       afero.NewMemMapFs(),
			OpenFunc: func(afero.Fs, string) (afero.File, error) { return nil, errors.New("open failed") },
		}
		path := filepath.Clean("/home/user/.wbctl/config.json")
		require.NoError(t, fsys.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, afero.WriteFile(fsys, path, []byte(`{}`), 0o644))
		doc, err := config.ReadDocument(fsys, path)
		require.EqualError(t, err, "Failed to read config file '"+path+"'\n"+
			" Caused by:\n"+
			"\t* open failed")
		require.Empty(t, doc)
	})

	t.Run("read of invalid file returns error", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		path := filepath.Clean("/home/user/.wbctl/config.json")
		require.NoError(t, fsys.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, afero.WriteFile(fsys, path, []byte(`invalid`), 0o644))
		doc, err := config.ReadDocument(fsys, path)
		require.EqualError(t, err, "Failed to read config file '"+path+"'\n"+
			" Caused by:\n"+
			"\t* invalid character 'i' looking for beginning of value")
		require.Empty(t, doc)
	})

	t.Run("read unsupported file format is returned", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		path := filepath.Clean("/home/user/.wbctl/config.toml")
		require.NoError(t, fsys.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, afero.WriteFile(fsys, path, []byte(`one = two`), 0o644))
		_, err := config.ReadDocument(fsys, path)
		require.EqualError(t, err, "Failed to read config file '"+path+"'\n"+
			" Caused by:\n"+
			"\t* Unsupported config file format for '"+path+"'; only JSON and YAML can be written")
	})
}

func TestWriteDocument(t *testing.T) {
	t.Run("write mkdir failure is returned", func(t *testing.T) {
		fsys := stubs.StubFS{
			Fs:           afero.NewMemMapFs(),
			MkdirAllFunc: func(afero.Fs, string, os.FileMode) error { return errors.New("mkdir failed") },
		}
		doc := map[string]any{"profile": map[string]any{"endpoint": "https://example.com"}}
		err := config.WriteDocument(fsys, "/home/user/.wbctl/config.json", doc)
		require.EqualError(t, err, "Failed to write config file '/home/user/.wbctl/config.json'\n"+
			" Caused by:\n"+
			"\t* mkdir failed")
	})

	t.Run("write marshal json failure", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		doc := map[string]any{"invalid": make(chan int)}
		err := config.WriteDocument(fsys, "/home/user/.wbctl/config.json", doc)
		require.EqualError(t, err, "Failed to write config file '/home/user/.wbctl/config.json'\n"+
			" Caused by:\n"+
			"\t* json: unsupported type: chan int")
	})

	t.Run("write writefile failure is returned", func(t *testing.T) {
		fsys := stubs.StubFS{
			Fs: afero.NewMemMapFs(),
			OpenFileFunc: func(afero.Fs, string, int, os.FileMode) (afero.File, error) {
				return nil, errors.New("openfile failed")
			},
		}
		doc := map[string]any{"profile": map[string]any{"endpoint": "https://example.com"}}
		err := config.WriteDocument(fsys, "/home/user/.wbctl/config.json", doc)
		require.EqualError(t, err, "Failed to write config file '/home/user/.wbctl/config.json'\n"+
			" Caused by:\n"+
			"\t* openfile failed")
	})

	t.Run("write chmod failure is returned", func(t *testing.T) {
		fsys := stubs.StubFS{
			Fs: afero.NewMemMapFs(),
			ChmodFunc: func(afero.Fs, string, os.FileMode) error {
				return errors.New("chmod failed")
			},
		}
		doc := map[string]any{"profile": map[string]any{"endpoint": "https://example.com"}}
		err := config.WriteDocument(fsys, "/home/user/.wbctl/config.json", doc)
		require.EqualError(t, err, "Failed to write config file '/home/user/.wbctl/config.json'\n"+
			" Caused by:\n"+
			"\t* chmod failed")
	})

	t.Run("write unsupported file format is returned", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		doc := map[string]any{}
		err := config.WriteDocument(fsys, "/home/user/.wbctl/config.toml", doc)
		require.EqualError(t, err, "Failed to write config file '/home/user/.wbctl/config.toml'\n"+
			" Caused by:\n"+
			"\t* Unsupported config file format for '/home/user/.wbctl/config.toml'; only JSON and YAML can be written")
	})
}
