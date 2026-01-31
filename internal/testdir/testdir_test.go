/*
 * Copyright 2026 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package testdir_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dmikusa/bptest/internal/testdir"
)

func TestManager_CreateTempDir(t *testing.T) {
	m := testdir.NewManager()
	defer m.Cleanup()

	dir, err := m.CreateTempDir("test-")
	if err != nil {
		t.Fatalf("CreateTempDir failed: %v", err)
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("expected temp dir to exist")
	}

	dirs := m.Dirs()
	if len(dirs) != 1 || dirs[0] != dir {
		t.Errorf("expected Dirs() to return [%s], got %v", dir, dirs)
	}
}

func TestManager_CreateMultipleTempDirs(t *testing.T) {
	m := testdir.NewManager()
	defer m.Cleanup()

	dir1, err := m.CreateTempDir("test1-")
	if err != nil {
		t.Fatalf("CreateTempDir failed: %v", err)
	}

	dir2, err := m.CreateTempDir("test2-")
	if err != nil {
		t.Fatalf("CreateTempDir failed: %v", err)
	}

	dirs := m.Dirs()
	if len(dirs) != 2 {
		t.Errorf("expected 2 dirs, got %d", len(dirs))
	}

	if dirs[0] != dir1 || dirs[1] != dir2 {
		t.Errorf("unexpected dirs: %v", dirs)
	}
}

func TestManager_WriteFile(t *testing.T) {
	m := testdir.NewManager()
	defer m.Cleanup()

	dir, err := m.CreateTempDir("test-")
	if err != nil {
		t.Fatalf("CreateTempDir failed: %v", err)
	}

	content := []byte("hello world")
	err = m.WriteFile(dir, "subdir/file.txt", content)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	fullPath := filepath.Join(dir, "subdir", "file.txt")
	readContent, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(readContent) != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", string(readContent))
	}
}

func TestManager_WriteFileString(t *testing.T) {
	m := testdir.NewManager()
	defer m.Cleanup()

	dir, err := m.CreateTempDir("test-")
	if err != nil {
		t.Fatalf("CreateTempDir failed: %v", err)
	}

	err = m.WriteFileString(dir, "test.txt", "string content")
	if err != nil {
		t.Fatalf("WriteFileString failed: %v", err)
	}

	fullPath := filepath.Join(dir, "test.txt")
	content, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(content) != "string content" {
		t.Errorf("expected 'string content', got '%s'", string(content))
	}
}

func TestManager_WriteFile_CreatesParentDirs(t *testing.T) {
	m := testdir.NewManager()
	defer m.Cleanup()

	dir, err := m.CreateTempDir("test-")
	if err != nil {
		t.Fatalf("CreateTempDir failed: %v", err)
	}

	err = m.WriteFile(dir, "a/b/c/deep.txt", []byte("deep"))
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	fullPath := filepath.Join(dir, "a", "b", "c", "deep.txt")
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Error("expected deep file to exist")
	}
}

func TestManager_Cleanup(t *testing.T) {
	m := testdir.NewManager()

	dir1, _ := m.CreateTempDir("test1-")
	dir2, _ := m.CreateTempDir("test2-")

	// Verify dirs exist
	if _, err := os.Stat(dir1); os.IsNotExist(err) {
		t.Fatal("dir1 should exist before cleanup")
	}
	if _, err := os.Stat(dir2); os.IsNotExist(err) {
		t.Fatal("dir2 should exist before cleanup")
	}

	err := m.Cleanup()
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify dirs are removed
	if _, err := os.Stat(dir1); !os.IsNotExist(err) {
		t.Error("dir1 should not exist after cleanup")
	}
	if _, err := os.Stat(dir2); !os.IsNotExist(err) {
		t.Error("dir2 should not exist after cleanup")
	}

	// Verify Dirs() is cleared
	if len(m.Dirs()) != 0 {
		t.Error("expected Dirs() to be empty after cleanup")
	}
}

func TestManager_Cleanup_EmptyManager(t *testing.T) {
	m := testdir.NewManager()

	err := m.Cleanup()
	if err != nil {
		t.Errorf("Cleanup on empty manager should not error: %v", err)
	}
}
