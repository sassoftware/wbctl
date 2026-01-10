// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stubs

import (
	"os"
	"time"

	"github.com/spf13/afero"
)

var _ afero.Fs = StubFS{}

// StubFS is a test filesystem that delegates to injected functions when set.
type StubFS struct {
	afero.Fs
	CreateFunc    func(afero.Fs, string) (afero.File, error)
	MkdirFunc     func(afero.Fs, string, os.FileMode) error
	MkdirAllFunc  func(afero.Fs, string, os.FileMode) error
	OpenFunc      func(afero.Fs, string) (afero.File, error)
	OpenFileFunc  func(afero.Fs, string, int, os.FileMode) (afero.File, error)
	RemoveFunc    func(afero.Fs, string) error
	RemoveAllFunc func(afero.Fs, string) error
	RenameFunc    func(afero.Fs, string, string) error
	StatFunc      func(afero.Fs, string) (os.FileInfo, error)
	NameFunc      func(afero.Fs) string
	ChmodFunc     func(afero.Fs, string, os.FileMode) error
	ChownFunc     func(afero.Fs, string, int, int) error
	ChtimesFunc   func(afero.Fs, string, time.Time, time.Time) error
}

// Create opens a file for writing, using the injected function when present.
func (s StubFS) Create(name string) (afero.File, error) {
	if s.CreateFunc != nil {
		return s.CreateFunc(s.Fs, name)
	}
	return s.Fs.Create(name)
}

// Mkdir creates a directory, using the injected function when present.
func (s StubFS) Mkdir(name string, perm os.FileMode) error {
	if s.MkdirFunc != nil {
		return s.MkdirFunc(s.Fs, name, perm)
	}
	return s.Fs.Mkdir(name, perm)
}

// MkdirAll creates a directory tree, using the injected function when present.
func (s StubFS) MkdirAll(path string, perm os.FileMode) error {
	if s.MkdirAllFunc != nil {
		return s.MkdirAllFunc(s.Fs, path, perm)
	}
	return s.Fs.MkdirAll(path, perm)
}

// Open opens a file, using the injected function when present.
func (s StubFS) Open(name string) (afero.File, error) {
	if s.OpenFunc != nil {
		return s.OpenFunc(s.Fs, name)
	}
	return s.Fs.Open(name)
}

// OpenFile opens a file with flags, using the injected function when present.
func (s StubFS) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if s.OpenFileFunc != nil {
		return s.OpenFileFunc(s.Fs, name, flag, perm)
	}
	return s.Fs.OpenFile(name, flag, perm)
}

// Remove deletes a file, using the injected function when present.
func (s StubFS) Remove(name string) error {
	if s.RemoveFunc != nil {
		return s.RemoveFunc(s.Fs, name)
	}
	return s.Fs.Remove(name)
}

// RemoveAll deletes a path tree, using the injected function when present.
func (s StubFS) RemoveAll(path string) error {
	if s.RemoveAllFunc != nil {
		return s.RemoveAllFunc(s.Fs, path)
	}
	return s.Fs.RemoveAll(path)
}

// Rename renames a file, using the injected function when present.
func (s StubFS) Rename(oldname, newname string) error {
	if s.RenameFunc != nil {
		return s.RenameFunc(s.Fs, oldname, newname)
	}
	return s.Fs.Rename(oldname, newname)
}

// Stat returns file information, using the injected function when present.
func (s StubFS) Stat(name string) (os.FileInfo, error) {
	if s.StatFunc != nil {
		return s.StatFunc(s.Fs, name)
	}
	return s.Fs.Stat(name)
}

// Name returns the filesystem name, using the injected function when present.
func (s StubFS) Name() string {
	if s.NameFunc != nil {
		return s.NameFunc(s.Fs)
	}
	return s.Fs.Name()
}

// Chmod changes file permissions, using the injected function when present.
func (s StubFS) Chmod(name string, mode os.FileMode) error {
	if s.ChmodFunc != nil {
		return s.ChmodFunc(s.Fs, name, mode)
	}
	return s.Fs.Chmod(name, mode)
}

// Chown changes file ownership, using the injected function when present.
func (s StubFS) Chown(name string, uid, gid int) error {
	if s.ChownFunc != nil {
		return s.ChownFunc(s.Fs, name, uid, gid)
	}
	return s.Fs.Chown(name, uid, gid)
}

// Chtimes updates file timestamps, using the injected function when present.
func (s StubFS) Chtimes(name string, atime time.Time, mtime time.Time) error {
	if s.ChtimesFunc != nil {
		return s.ChtimesFunc(s.Fs, name, atime, mtime)
	}
	return s.Fs.Chtimes(name, atime, mtime)
}

var _ afero.File = StubFile{}

// StubFile is a test file that delegates to injected functions when set.
type StubFile struct {
	afero.File
	CloseFunc        func(afero.File) error
	ReadFunc         func(afero.File, []byte) (int, error)
	ReadAtFunc       func(afero.File, []byte, int64) (int, error)
	SeekFunc         func(afero.File, int64, int) (int64, error)
	WriteFunc        func(afero.File, []byte) (int, error)
	WriteAtFunc      func(afero.File, []byte, int64) (int, error)
	NameFunc         func(afero.File) string
	ReaddirFunc      func(afero.File, int) ([]os.FileInfo, error)
	ReaddirnamesFunc func(afero.File, int) ([]string, error)
	StatFunc         func(afero.File) (os.FileInfo, error)
	SyncFunc         func(afero.File) error
	TruncateFunc     func(afero.File, int64) error
	WriteStringFunc  func(afero.File, string) (int, error)
}

// Close closes the file, using the injected function when present.
func (s StubFile) Close() error {
	if s.CloseFunc != nil {
		return s.CloseFunc(s.File)
	}
	return s.File.Close()
}

// Read reads from the file, using the injected function when present.
func (s StubFile) Read(p []byte) (int, error) {
	if s.ReadFunc != nil {
		return s.ReadFunc(s.File, p)
	}
	return s.File.Read(p)
}

// ReadAt reads from an offset, using the injected function when present.
func (s StubFile) ReadAt(p []byte, off int64) (int, error) {
	if s.ReadAtFunc != nil {
		return s.ReadAtFunc(s.File, p, off)
	}
	return s.File.ReadAt(p, off)
}

// Seek changes the file offset, using the injected function when present.
func (s StubFile) Seek(offset int64, whence int) (int64, error) {
	if s.SeekFunc != nil {
		return s.SeekFunc(s.File, offset, whence)
	}
	return s.File.Seek(offset, whence)
}

// Write writes to the file, using the injected function when present.
func (s StubFile) Write(p []byte) (int, error) {
	if s.WriteFunc != nil {
		return s.WriteFunc(s.File, p)
	}
	return s.File.Write(p)
}

// WriteAt writes at an offset, using the injected function when present.
func (s StubFile) WriteAt(p []byte, off int64) (int, error) {
	if s.WriteAtFunc != nil {
		return s.WriteAtFunc(s.File, p, off)
	}
	return s.File.WriteAt(p, off)
}

// Name returns the file name, using the injected function when present.
func (s StubFile) Name() string {
	if s.NameFunc != nil {
		return s.NameFunc(s.File)
	}
	return s.File.Name()
}

// Readdir reads directory entries, using the injected function when present.
func (s StubFile) Readdir(count int) ([]os.FileInfo, error) {
	if s.ReaddirFunc != nil {
		return s.ReaddirFunc(s.File, count)
	}
	return s.File.Readdir(count)
}

// Readdirnames reads directory entry names, using the injected function when present.
func (s StubFile) Readdirnames(n int) ([]string, error) {
	if s.ReaddirnamesFunc != nil {
		return s.ReaddirnamesFunc(s.File, n)
	}
	return s.File.Readdirnames(n)
}

// Stat returns file information, using the injected function when present.
func (s StubFile) Stat() (os.FileInfo, error) {
	if s.StatFunc != nil {
		return s.StatFunc(s.File)
	}
	return s.File.Stat()
}

// Sync flushes file contents, using the injected function when present.
func (s StubFile) Sync() error {
	if s.SyncFunc != nil {
		return s.SyncFunc(s.File)
	}
	return s.File.Sync()
}

// Truncate truncates the file, using the injected function when present.
func (s StubFile) Truncate(size int64) error {
	if s.TruncateFunc != nil {
		return s.TruncateFunc(s.File, size)
	}
	return s.File.Truncate(size)
}

// WriteString writes a string, using the injected function when present.
func (s StubFile) WriteString(st string) (int, error) {
	if s.WriteStringFunc != nil {
		return s.WriteStringFunc(s.File, st)
	}
	return s.File.WriteString(st)
}
