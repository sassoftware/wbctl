// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stubs_test

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ---- StubFS ----------------------------------------------------------------

func TestStubFS_Create(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		base := afero.NewMemMapFs()
		fsys := stubs.StubFS{Fs: base}

		f, err := fsys.Create("file.txt")
		require.NoError(t, err)
		require.NoError(t, f.Close())
		_, err = fsys.Stat("file.txt")
		require.NoError(t, err)
	})

	t.Run("hook", func(t *testing.T) {
		base := afero.NewMemMapFs()
		called := false
		fsys := stubs.StubFS{
			Fs: base,
			CreateFunc: func(_ afero.Fs, name string) (afero.File, error) {
				called = true
				require.Equal(t, "file.txt", name)
				return base.Create("hook-" + name)
			},
		}

		f, err := fsys.Create("file.txt")
		require.NoError(t, err)
		require.True(t, called)
		require.NoError(t, f.Close())
		_, err = fsys.Stat("hook-file.txt")
		require.NoError(t, err)
	})
}

func TestStubFS_Mkdir(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := stubs.StubFS{Fs: afero.NewMemMapFs()}
		require.NoError(t, fsys.Mkdir("dir", 0o700))
		info, err := fsys.Stat("dir")
		require.NoError(t, err)
		require.True(t, info.IsDir())
	})

	t.Run("hook", func(t *testing.T) {
		base := afero.NewMemMapFs()
		called := false
		wantErr := errors.New("mkdir hook")
		fsys := stubs.StubFS{
			Fs: base,
			MkdirFunc: func(_ afero.Fs, name string, perm os.FileMode) error {
				called = true
				require.Equal(t, "dir", name)
				require.Equal(t, os.FileMode(0o700), perm)
				return wantErr
			},
		}

		require.ErrorIs(t, fsys.Mkdir("dir", 0o700), wantErr)
		require.True(t, called)
	})
}

func TestStubFS_MkdirAll(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := stubs.StubFS{Fs: afero.NewMemMapFs()}
		require.NoError(t, fsys.MkdirAll(filepath.Join("a", "b"), 0o700))
		info, err := fsys.Stat(filepath.Join("a", "b"))
		require.NoError(t, err)
		require.True(t, info.IsDir())
	})

	t.Run("hook", func(t *testing.T) {
		base := afero.NewMemMapFs()
		called := false
		wantErr := errors.New("mkdirall hook")
		fsys := stubs.StubFS{
			Fs: base,
			MkdirAllFunc: func(_ afero.Fs, path string, perm os.FileMode) error {
				called = true
				require.Equal(t, filepath.Join("a", "b"), path)
				require.Equal(t, os.FileMode(0o700), perm)
				return wantErr
			},
		}

		require.ErrorIs(t, fsys.MkdirAll(filepath.Join("a", "b"), 0o700), wantErr)
		require.True(t, called)
	})
}

func TestStubFS_Open(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		base := afero.NewMemMapFs()
		require.NoError(t, afero.WriteFile(base, "file.txt", []byte("hello"), 0o600))
		fsys := stubs.StubFS{Fs: base}

		f, err := fsys.Open("file.txt")
		require.NoError(t, err)
		defer func() { require.NoError(t, f.Close()) }()

		buf := make([]byte, 5)
		n, err := f.Read(buf)
		require.NoError(t, err)
		require.Equal(t, 5, n)
		require.Equal(t, "hello", string(buf))
	})

	t.Run("hook", func(t *testing.T) {
		base := afero.NewMemMapFs()
		require.NoError(t, afero.WriteFile(base, "hooked.txt", []byte("hook"), 0o600))
		called := false
		fsys := stubs.StubFS{
			Fs: base,
			OpenFunc: func(_ afero.Fs, name string) (afero.File, error) {
				called = true
				require.Equal(t, "hooked.txt", name)
				return base.Open(name)
			},
		}

		f, err := fsys.Open("hooked.txt")
		require.NoError(t, err)
		require.True(t, called)
		defer func() { require.NoError(t, f.Close()) }()
		buf := make([]byte, 4)
		n, err := f.Read(buf)
		require.NoError(t, err)
		require.Equal(t, 4, n)
		require.Equal(t, "hook", string(buf))
	})
}

func TestStubFS_OpenFile(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := stubs.StubFS{Fs: afero.NewMemMapFs()}
		f, err := fsys.OpenFile("file.txt", os.O_CREATE|os.O_RDWR, 0o600)
		require.NoError(t, err)
		_, err = f.WriteString("hello")
		require.NoError(t, err)
		require.NoError(t, f.Close())

		data, err := afero.ReadFile(fsys, "file.txt")
		require.NoError(t, err)
		require.Equal(t, "hello", string(data))
	})

	t.Run("hook", func(t *testing.T) {
		base := afero.NewMemMapFs()
		called := false
		fsys := stubs.StubFS{
			Fs: base,
			OpenFileFunc: func(_ afero.Fs, name string, flag int, perm os.FileMode) (afero.File, error) {
				called = true
				require.Equal(t, "hooked.txt", name)
				require.Equal(t, os.O_CREATE|os.O_RDWR, flag)
				require.Equal(t, os.FileMode(0o600), perm)
				return base.OpenFile("hooked.txt", flag, perm)
			},
		}

		f, err := fsys.OpenFile("hooked.txt", os.O_CREATE|os.O_RDWR, 0o600)
		require.NoError(t, err)
		require.True(t, called)
		_, err = f.WriteString("hook")
		require.NoError(t, err)
		require.NoError(t, f.Close())

		data, err := afero.ReadFile(base, "hooked.txt")
		require.NoError(t, err)
		require.Equal(t, "hook", string(data))
	})
}

func TestStubFS_Remove(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := stubs.StubFS{Fs: afero.NewMemMapFs()}
		require.NoError(t, afero.WriteFile(fsys, "file.txt", []byte("hello"), 0o600))
		require.NoError(t, fsys.Remove("file.txt"))
		_, err := fsys.Stat("file.txt")
		require.Error(t, err)
	})

	t.Run("hook", func(t *testing.T) {
		base := afero.NewMemMapFs()
		called := false
		wantErr := errors.New("remove hook")
		fsys := stubs.StubFS{
			Fs: base,
			RemoveFunc: func(_ afero.Fs, name string) error {
				called = true
				require.Equal(t, "file.txt", name)
				return wantErr
			},
		}

		require.ErrorIs(t, fsys.Remove("file.txt"), wantErr)
		require.True(t, called)
	})
}

func TestStubFS_RemoveAll(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := stubs.StubFS{Fs: afero.NewMemMapFs()}
		require.NoError(t, fsys.MkdirAll(filepath.Join("a", "b"), 0o700))
		require.NoError(t, afero.WriteFile(fsys, filepath.Join("a", "b", "file.txt"), []byte("hello"), 0o600))
		require.NoError(t, fsys.RemoveAll("a"))
		_, err := fsys.Stat("a")
		require.Error(t, err)
	})

	t.Run("hook", func(t *testing.T) {
		base := afero.NewMemMapFs()
		called := false
		wantErr := errors.New("removeall hook")
		fsys := stubs.StubFS{
			Fs: base,
			RemoveAllFunc: func(_ afero.Fs, path string) error {
				called = true
				require.Equal(t, "a", path)
				return wantErr
			},
		}

		require.ErrorIs(t, fsys.RemoveAll("a"), wantErr)
		require.True(t, called)
	})
}

func TestStubFS_Rename(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := stubs.StubFS{Fs: afero.NewMemMapFs()}
		require.NoError(t, afero.WriteFile(fsys, "old.txt", []byte("hello"), 0o600))
		require.NoError(t, fsys.Rename("old.txt", "new.txt"))
		_, err := fsys.Stat("old.txt")
		require.Error(t, err)
		data, err := afero.ReadFile(fsys, "new.txt")
		require.NoError(t, err)
		require.Equal(t, "hello", string(data))
	})

	t.Run("hook", func(t *testing.T) {
		base := afero.NewMemMapFs()
		called := false
		wantErr := errors.New("rename hook")
		fsys := stubs.StubFS{
			Fs: base,
			RenameFunc: func(_ afero.Fs, oldname, newname string) error {
				called = true
				require.Equal(t, "old.txt", oldname)
				require.Equal(t, "new.txt", newname)
				return wantErr
			},
		}

		require.ErrorIs(t, fsys.Rename("old.txt", "new.txt"), wantErr)
		require.True(t, called)
	})
}

func TestStubFS_Stat(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := stubs.StubFS{Fs: afero.NewMemMapFs()}
		require.NoError(t, afero.WriteFile(fsys, "file.txt", []byte("hello"), 0o600))
		info, err := fsys.Stat("file.txt")
		require.NoError(t, err)
		require.Equal(t, int64(5), info.Size())
	})

	t.Run("hook", func(t *testing.T) {
		base := afero.NewMemMapFs()
		require.NoError(t, afero.WriteFile(base, "hooked.txt", []byte("hook"), 0o600))
		called := false
		fsys := stubs.StubFS{
			Fs: base,
			StatFunc: func(_ afero.Fs, name string) (os.FileInfo, error) {
				called = true
				require.Equal(t, "hooked.txt", name)
				return base.Stat(name)
			},
		}

		info, err := fsys.Stat("hooked.txt")
		require.NoError(t, err)
		require.True(t, called)
		require.Equal(t, int64(4), info.Size())
	})
}

func TestStubFS_Name(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := stubs.StubFS{Fs: afero.NewMemMapFs()}
		require.Equal(t, "MemMapFS", fsys.Name())
	})

	t.Run("hook", func(t *testing.T) {
		base := afero.NewMemMapFs()
		called := false
		fsys := stubs.StubFS{
			Fs: base,
			NameFunc: func(_ afero.Fs) string {
				called = true
				return "hooked"
			},
		}

		require.Equal(t, "hooked", fsys.Name())
		require.True(t, called)
	})
}

func TestStubFS_Chmod(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := stubs.StubFS{Fs: afero.NewMemMapFs()}
		require.NoError(t, afero.WriteFile(fsys, "file.txt", []byte("hello"), 0o600))
		require.NoError(t, fsys.Chmod("file.txt", 0o644))
	})

	t.Run("hook", func(t *testing.T) {
		base := afero.NewMemMapFs()
		called := false
		wantErr := errors.New("chmod hook")
		fsys := stubs.StubFS{
			Fs: base,
			ChmodFunc: func(_ afero.Fs, name string, mode os.FileMode) error {
				called = true
				require.Equal(t, "file.txt", name)
				require.Equal(t, os.FileMode(0o644), mode)
				return wantErr
			},
		}

		require.ErrorIs(t, fsys.Chmod("file.txt", 0o644), wantErr)
		require.True(t, called)
	})
}

func TestStubFS_Chown(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		base := afero.NewMemMapFs()
		stub := stubs.StubFS{Fs: afero.NewMemMapFs()}
		require.NoError(t, afero.WriteFile(base, "file.txt", []byte("hello"), 0o600))
		require.NoError(t, afero.WriteFile(stub, "file.txt", []byte("hello"), 0o600))

		require.Equal(t, base.Chown("file.txt", 1, 2), stub.Chown("file.txt", 1, 2))
	})

	t.Run("hook", func(t *testing.T) {
		base := afero.NewMemMapFs()
		called := false
		wantErr := errors.New("chown hook")
		fsys := stubs.StubFS{
			Fs: base,
			ChownFunc: func(_ afero.Fs, name string, uid, gid int) error {
				called = true
				require.Equal(t, "file.txt", name)
				require.Equal(t, 1, uid)
				require.Equal(t, 2, gid)
				return wantErr
			},
		}

		require.ErrorIs(t, fsys.Chown("file.txt", 1, 2), wantErr)
		require.True(t, called)
	})
}

func TestStubFS_Chtimes(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		when := time.Unix(123, 0)
		base := afero.NewMemMapFs()
		stub := stubs.StubFS{Fs: afero.NewMemMapFs()}
		require.NoError(t, afero.WriteFile(base, "file.txt", []byte("hello"), 0o600))
		require.NoError(t, afero.WriteFile(stub, "file.txt", []byte("hello"), 0o600))

		require.Equal(t, base.Chtimes("file.txt", when, when), stub.Chtimes("file.txt", when, when))
	})

	t.Run("hook", func(t *testing.T) {
		when := time.Unix(123, 0)
		base := afero.NewMemMapFs()
		called := false
		wantErr := errors.New("chtimes hook")
		fsys := stubs.StubFS{
			Fs: base,
			ChtimesFunc: func(_ afero.Fs, name string, atime, mtime time.Time) error {
				called = true
				require.Equal(t, "file.txt", name)
				require.Equal(t, when, atime)
				require.Equal(t, when, mtime)
				return wantErr
			},
		}

		require.ErrorIs(t, fsys.Chtimes("file.txt", when, when), wantErr)
		require.True(t, called)
	})
}

// ---- StubFile --------------------------------------------------------------

func TestStubFile_Close(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		f, err := fsys.Create("file.txt")
		require.NoError(t, err)
		stub := stubs.StubFile{File: f}
		require.NoError(t, stub.Close())
	})

	t.Run("hook", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		f, err := fsys.Create("file.txt")
		require.NoError(t, err)
		called := false
		wantErr := errors.New("close hook")
		stub := stubs.StubFile{
			File: f,
			CloseFunc: func(_ afero.File) error {
				called = true
				return wantErr
			},
		}

		require.ErrorIs(t, stub.Close(), wantErr)
		require.True(t, called)
	})
}

func TestStubFile_Read(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		require.NoError(t, afero.WriteFile(fsys, "file.txt", []byte("hello"), 0o600))
		f, err := fsys.Open("file.txt")
		require.NoError(t, err)
		defer func() { require.NoError(t, f.Close()) }()

		stub := stubs.StubFile{File: f}
		buf := make([]byte, 5)
		n, err := stub.Read(buf)
		require.NoError(t, err)
		require.Equal(t, 5, n)
		require.Equal(t, "hello", string(buf))
	})

	t.Run("hook", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		require.NoError(t, afero.WriteFile(fsys, "file.txt", []byte("hello"), 0o600))
		f, err := fsys.Open("file.txt")
		require.NoError(t, err)
		defer func() { require.NoError(t, f.Close()) }()

		called := false
		stub := stubs.StubFile{
			File: f,
			ReadFunc: func(_ afero.File, p []byte) (int, error) {
				called = true
				copy(p, []byte("hook"))
				return 4, nil
			},
		}

		buf := make([]byte, 5)
		n, err := stub.Read(buf)
		require.NoError(t, err)
		require.True(t, called)
		require.Equal(t, 4, n)
		require.Equal(t, "hook\x00", string(buf))
	})
}

func TestStubFile_ReadAt(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		require.NoError(t, afero.WriteFile(fsys, "file.txt", []byte("hello"), 0o600))
		f, err := fsys.Open("file.txt")
		require.NoError(t, err)
		defer func() { require.NoError(t, f.Close()) }()

		stub := stubs.StubFile{File: f}
		buf := make([]byte, 3)
		n, err := stub.ReadAt(buf, 1)
		require.NoError(t, err)
		require.Equal(t, 3, n)
		require.Equal(t, "ell", string(buf))
	})

	t.Run("hook", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		require.NoError(t, afero.WriteFile(fsys, "file.txt", []byte("hello"), 0o600))
		f, err := fsys.Open("file.txt")
		require.NoError(t, err)
		defer func() { require.NoError(t, f.Close()) }()

		called := false
		stub := stubs.StubFile{
			File: f,
			ReadAtFunc: func(_ afero.File, p []byte, off int64) (int, error) {
				called = true
				require.Equal(t, int64(1), off)
				copy(p, []byte("abc"))
				return 3, nil
			},
		}

		buf := make([]byte, 3)
		n, err := stub.ReadAt(buf, 1)
		require.NoError(t, err)
		require.True(t, called)
		require.Equal(t, 3, n)
		require.Equal(t, "abc", string(buf))
	})
}

func TestStubFile_Seek(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		require.NoError(t, afero.WriteFile(fsys, "file.txt", []byte("hello"), 0o600))
		f, err := fsys.Open("file.txt")
		require.NoError(t, err)
		defer func() { require.NoError(t, f.Close()) }()

		stub := stubs.StubFile{File: f}
		got, err := stub.Seek(2, io.SeekStart)
		require.NoError(t, err)
		require.Equal(t, int64(2), got)
	})

	t.Run("hook", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		require.NoError(t, afero.WriteFile(fsys, "file.txt", []byte("hello"), 0o600))
		f, err := fsys.Open("file.txt")
		require.NoError(t, err)
		defer func() { require.NoError(t, f.Close()) }()

		called := false
		stub := stubs.StubFile{
			File: f,
			SeekFunc: func(_ afero.File, offset int64, whence int) (int64, error) {
				called = true
				require.Equal(t, int64(2), offset)
				require.Equal(t, io.SeekStart, whence)
				return 7, nil
			},
		}

		got, err := stub.Seek(2, io.SeekStart)
		require.NoError(t, err)
		require.True(t, called)
		require.Equal(t, int64(7), got)
	})
}

func TestStubFile_Write(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		f, err := fsys.OpenFile("file.txt", os.O_CREATE|os.O_RDWR, 0o600)
		require.NoError(t, err)
		stub := stubs.StubFile{File: f}

		n, err := stub.Write([]byte("hello"))
		require.NoError(t, err)
		require.Equal(t, 5, n)
		require.NoError(t, stub.Close())

		data, err := afero.ReadFile(fsys, "file.txt")
		require.NoError(t, err)
		require.Equal(t, "hello", string(data))
	})

	t.Run("hook", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		f, err := fsys.OpenFile("file.txt", os.O_CREATE|os.O_RDWR, 0o600)
		require.NoError(t, err)
		stub := stubs.StubFile{
			File: f,
			WriteFunc: func(_ afero.File, p []byte) (int, error) {
				require.Equal(t, []byte("hello"), p)
				return 3, nil
			},
		}

		n, err := stub.Write([]byte("hello"))
		require.NoError(t, err)
		require.Equal(t, 3, n)
		require.NoError(t, stub.Close())
	})
}

func TestStubFile_WriteAt(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		f, err := fsys.OpenFile("file.txt", os.O_CREATE|os.O_RDWR, 0o600)
		require.NoError(t, err)
		stub := stubs.StubFile{File: f}

		n, err := stub.WriteAt([]byte("hello"), 0)
		require.NoError(t, err)
		require.Equal(t, 5, n)
		require.NoError(t, stub.Close())

		data, err := afero.ReadFile(fsys, "file.txt")
		require.NoError(t, err)
		require.Equal(t, "hello", string(data))
	})

	t.Run("hook", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		f, err := fsys.OpenFile("file.txt", os.O_CREATE|os.O_RDWR, 0o600)
		require.NoError(t, err)
		stub := stubs.StubFile{
			File: f,
			WriteAtFunc: func(_ afero.File, p []byte, off int64) (int, error) {
				require.Equal(t, []byte("hello"), p)
				require.Equal(t, int64(0), off)
				return 3, nil
			},
		}

		n, err := stub.WriteAt([]byte("hello"), 0)
		require.NoError(t, err)
		require.Equal(t, 3, n)
		require.NoError(t, stub.Close())
	})
}

func TestStubFile_Name(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		f, err := fsys.Create("file.txt")
		require.NoError(t, err)
		defer func() { require.NoError(t, f.Close()) }()

		stub := stubs.StubFile{File: f}
		require.NotEmpty(t, stub.Name())
	})

	t.Run("hook", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		f, err := fsys.Create("file.txt")
		require.NoError(t, err)
		defer func() { require.NoError(t, f.Close()) }()

		called := false
		stub := stubs.StubFile{
			File: f,
			NameFunc: func(_ afero.File) string {
				called = true
				return "hooked.txt"
			},
		}

		require.Equal(t, "hooked.txt", stub.Name())
		require.True(t, called)
	})
}

func TestStubFile_Readdir(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		require.NoError(t, fsys.MkdirAll("dir", 0o700))
		require.NoError(t, afero.WriteFile(fsys, filepath.Join("dir", "a.txt"), []byte("a"), 0o600))
		dir, err := fsys.Open("dir")
		require.NoError(t, err)
		defer func() { require.NoError(t, dir.Close()) }()

		stub := stubs.StubFile{File: dir}
		infos, err := stub.Readdir(1)
		require.NoError(t, err)
		require.Len(t, infos, 1)
	})

	t.Run("hook", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		require.NoError(t, fsys.MkdirAll("dir", 0o700))
		dir, err := fsys.Open("dir")
		require.NoError(t, err)
		defer func() { require.NoError(t, dir.Close()) }()

		called := false
		stub := stubs.StubFile{
			File: dir,
			ReaddirFunc: func(_ afero.File, count int) ([]os.FileInfo, error) {
				called = true
				require.Equal(t, 1, count)
				return nil, nil
			},
		}

		infos, err := stub.Readdir(1)
		require.NoError(t, err)
		require.True(t, called)
		require.Nil(t, infos)
	})
}

func TestStubFile_Readdirnames(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		require.NoError(t, fsys.MkdirAll("dir", 0o700))
		require.NoError(t, afero.WriteFile(fsys, filepath.Join("dir", "a.txt"), []byte("a"), 0o600))
		dir, err := fsys.Open("dir")
		require.NoError(t, err)
		defer func() { require.NoError(t, dir.Close()) }()

		stub := stubs.StubFile{File: dir}
		names, err := stub.Readdirnames(1)
		require.NoError(t, err)
		require.Equal(t, []string{"a.txt"}, names)
	})

	t.Run("hook", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		require.NoError(t, fsys.MkdirAll("dir", 0o700))
		dir, err := fsys.Open("dir")
		require.NoError(t, err)
		defer func() { require.NoError(t, dir.Close()) }()

		called := false
		stub := stubs.StubFile{
			File: dir,
			ReaddirnamesFunc: func(_ afero.File, count int) ([]string, error) {
				called = true
				require.Equal(t, 1, count)
				return []string{"hooked.txt"}, nil
			},
		}

		names, err := stub.Readdirnames(1)
		require.NoError(t, err)
		require.True(t, called)
		require.Equal(t, []string{"hooked.txt"}, names)
	})
}

func TestStubFile_Stat(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		require.NoError(t, afero.WriteFile(fsys, "file.txt", []byte("hello"), 0o600))
		f, err := fsys.Open("file.txt")
		require.NoError(t, err)
		defer func() { require.NoError(t, f.Close()) }()

		stub := stubs.StubFile{File: f}
		info, err := stub.Stat()
		require.NoError(t, err)
		require.Equal(t, int64(5), info.Size())
	})

	t.Run("hook", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		require.NoError(t, afero.WriteFile(fsys, "file.txt", []byte("hello"), 0o600))
		f, err := fsys.Open("file.txt")
		require.NoError(t, err)
		defer func() { require.NoError(t, f.Close()) }()

		called := false
		stub := stubs.StubFile{
			File: f,
			StatFunc: func(_ afero.File) (os.FileInfo, error) {
				called = true
				return f.Stat()
			},
		}

		info, err := stub.Stat()
		require.NoError(t, err)
		require.True(t, called)
		require.Equal(t, int64(5), info.Size())
	})
}

func TestStubFile_Sync(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		f, err := fsys.OpenFile("file.txt", os.O_CREATE|os.O_RDWR, 0o600)
		require.NoError(t, err)
		defer func() { require.NoError(t, f.Close()) }()

		stub := stubs.StubFile{File: f}
		require.NoError(t, stub.Sync())
	})

	t.Run("hook", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		f, err := fsys.OpenFile("file.txt", os.O_CREATE|os.O_RDWR, 0o600)
		require.NoError(t, err)
		defer func() { require.NoError(t, f.Close()) }()

		called := false
		wantErr := errors.New("sync hook")
		stub := stubs.StubFile{
			File: f,
			SyncFunc: func(_ afero.File) error {
				called = true
				return wantErr
			},
		}

		require.ErrorIs(t, stub.Sync(), wantErr)
		require.True(t, called)
	})
}

func TestStubFile_Truncate(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		f, err := fsys.OpenFile("file.txt", os.O_CREATE|os.O_RDWR, 0o600)
		require.NoError(t, err)
		_, err = f.WriteString("hello")
		require.NoError(t, err)

		stub := stubs.StubFile{File: f}
		require.NoError(t, stub.Truncate(2))
		require.NoError(t, f.Close())

		data, err := afero.ReadFile(fsys, "file.txt")
		require.NoError(t, err)
		require.Equal(t, "he", string(data))
	})

	t.Run("hook", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		f, err := fsys.OpenFile("file.txt", os.O_CREATE|os.O_RDWR, 0o600)
		require.NoError(t, err)
		stub := stubs.StubFile{
			File: f,
			TruncateFunc: func(_ afero.File, size int64) error {
				require.Equal(t, int64(2), size)
				return nil
			},
		}

		require.NoError(t, stub.Truncate(2))
		require.NoError(t, stub.Close())
	})
}

func TestStubFile_WriteString(t *testing.T) {
	t.Run("delegate", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		f, err := fsys.OpenFile("file.txt", os.O_CREATE|os.O_RDWR, 0o600)
		require.NoError(t, err)
		stub := stubs.StubFile{File: f}

		n, err := stub.WriteString("hello")
		require.NoError(t, err)
		require.Equal(t, 5, n)
		require.NoError(t, stub.Close())

		data, err := afero.ReadFile(fsys, "file.txt")
		require.NoError(t, err)
		require.Equal(t, "hello", string(data))
	})

	t.Run("hook", func(t *testing.T) {
		fsys := afero.NewMemMapFs()
		f, err := fsys.OpenFile("file.txt", os.O_CREATE|os.O_RDWR, 0o600)
		require.NoError(t, err)
		stub := stubs.StubFile{
			File: f,
			WriteStringFunc: func(_ afero.File, st string) (int, error) {
				require.Equal(t, "hello", st)
				return 3, nil
			},
		}

		n, err := stub.WriteString("hello")
		require.NoError(t, err)
		require.Equal(t, 3, n)
		require.NoError(t, stub.Close())
	})
}
