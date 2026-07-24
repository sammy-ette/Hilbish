//go:build !midnight

package moonlight

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func withScratchCacheDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	prevOverride := cacheDirOverride
	prevPath := cacheDirPath
	cacheDirOverride = dir
	cacheDirOnce = new(sync.Once)

	prevFingerprintOverride := binaryFingerprintOverride
	binaryFingerprintOverride = 1
	binaryFingerprintOnce = new(sync.Once)

	t.Cleanup(func() {
		cacheDirOverride = prevOverride
		cacheDirOnce = new(sync.Once)
		cacheDirPath = prevPath

		binaryFingerprintOverride = prevFingerprintOverride
		binaryFingerprintOnce = new(sync.Once)
	})
}

func TestDoFileUsesBytecodeCacheOnSecondLoad(t *testing.T) {
	withScratchCacheDir(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "script.lua")
	if err := os.WriteFile(path, []byte("hilbish_cache_test_marker = 1 + 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewRuntime()

	if err := r.DoFile(path); err != nil {
		t.Fatalf("first DoFile: %v", err)
	}

	cachePath, err := cacheFilePath(path)
	if err != nil {
		t.Fatalf("cacheFileFor: %v", err)
	}
	if _, err := os.Stat(cachePath); err != nil {
		t.Fatalf("expected cache file to exist after first load: %v", err)
	}

	if err := r.DoFile(path); err != nil {
		t.Fatalf("second DoFile (cache hit): %v", err)
	}
	v, _ := r.DoString("return hilbish_cache_test_marker")
	if got, ok := v.TryInt(); !ok || got != 2 {
		t.Fatalf("expected marker 2 after cache-hit load, got %v", v)
	}
}

func TestDoFileCacheInvalidatesOnSourceChange(t *testing.T) {
	withScratchCacheDir(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "script.lua")
	if err := os.WriteFile(path, []byte("hilbish_cache_test_marker = 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewRuntime()
	if err := r.DoFile(path); err != nil {
		t.Fatalf("first DoFile: %v", err)
	}
	v, _ := r.DoString("return hilbish_cache_test_marker")
	if got, ok := v.TryInt(); !ok || got != 1 {
		t.Fatalf("expected marker 1 after first load, got %v", v)
	}

	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(path, []byte("hilbish_cache_test_marker = 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := r.DoFile(path); err != nil {
		t.Fatalf("second DoFile (source changed): %v", err)
	}
	v, _ = r.DoString("return hilbish_cache_test_marker")
	if got, ok := v.TryInt(); !ok || got != 2 {
		t.Fatalf("expected marker 2 after source change, got %v", v)
	}
}

func TestDoFileToleratesCorruptCache(t *testing.T) {
	withScratchCacheDir(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "script.lua")
	if err := os.WriteFile(path, []byte("hilbish_cache_test_marker = 42\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewRuntime()
	if err := r.DoFile(path); err != nil {
		t.Fatalf("first DoFile: %v", err)
	}

	cachePath, err := cacheFilePath(path)
	if err != nil {
		t.Fatalf("cacheFileFor: %v", err)
	}
	if err := os.WriteFile(cachePath, []byte("not a valid cache entry"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := r.DoFile(path); err != nil {
		t.Fatalf("DoFile with corrupt cache should fall back to source: %v", err)
	}
	v, _ := r.DoString("return hilbish_cache_test_marker")
	if got, ok := v.TryInt(); !ok || got != 42 {
		t.Fatalf("expected marker 42 after recompiling past corrupt cache, got %v", v)
	}
}

func TestDoFileIgnoresCacheFromDifferentBinary(t *testing.T) {
	withScratchCacheDir(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "script.lua")
	if err := os.WriteFile(path, []byte("hilbish_cache_test_marker = 11\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewRuntime()
	if err := r.DoFile(path); err != nil {
		t.Fatalf("first DoFile: %v", err)
	}

	binaryFingerprintOverride = 2
	binaryFingerprintOnce = new(sync.Once)

	if err := r.DoFile(path); err != nil {
		t.Fatalf("DoFile: %v", err)
	}
	v, _ := r.DoString("return hilbish_cache_test_marker")
	if got, ok := v.TryInt(); !ok || got != 11 {
		t.Fatalf("expected marker 11 after fingerprint-mismatch recompile, got %v", v)
	}
}

func TestLuaDofileUsesBytecodeCache(t *testing.T) {
	withScratchCacheDir(t)

	dir := t.TempDir()
	path := filepath.Join(dir, "script.lua")
	if err := os.WriteFile(path, []byte("hilbish_cache_test_marker = 7\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewRuntime()

	quoted := "'" + path + "'"
	if _, err := r.DoString("dofile(" + quoted + ")"); err != nil {
		t.Fatalf("first dofile(): %v", err)
	}

	cachePath, err := cacheFilePath(path)
	if err != nil {
		t.Fatalf("cacheFileFor: %v", err)
	}
	if _, err := os.Stat(cachePath); err != nil {
		t.Fatalf("expected cache file to exist after Lua-level dofile(): %v", err)
	}

	if _, err := r.DoString("hilbish_cache_test_marker = 0 dofile(" + quoted + ")"); err != nil {
		t.Fatalf("second dofile() (cache hit): %v", err)
	}
	v, _ := r.DoString("return hilbish_cache_test_marker")
	if got, ok := v.TryInt(); !ok || got != 7 {
		t.Fatalf("expected marker 7 after Lua-level dofile() cache hit, got %v", v)
	}
}

func TestLuaRequireUsesBytecodeCache(t *testing.T) {
	withScratchCacheDir(t)

	dir := t.TempDir()
	modPath := filepath.Join(dir, "hilbish_cache_test_mod.lua")
	if err := os.WriteFile(modPath, []byte("return { marker = 9 }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewRuntime()

	setPath := "package.path = package.path .. ';" + filepath.Join(dir, "?.lua") + "'"
	if _, err := r.DoString(setPath); err != nil {
		t.Fatalf("setting package.path: %v", err)
	}

	if _, err := r.DoString("return require('hilbish_cache_test_mod').marker"); err != nil {
		t.Fatalf("first require(): %v", err)
	}

	cachePath, err := cacheFilePath(modPath)
	if err != nil {
		t.Fatalf("cacheFileFor: %v", err)
	}
	if _, err := os.Stat(cachePath); err != nil {
		t.Fatalf("expected cache file to exist after require(): %v", err)
	}

	v, err := r.DoString("return dofile('" + modPath + "').marker")
	if err != nil {
		t.Fatalf("dofile of the same module path (cache hit): %v", err)
	}
	if got, ok := v.TryInt(); !ok || got != 9 {
		t.Fatalf("expected marker 9 after require()-then-dofile() cache hit, got %v", v)
	}
}

func TestDisableBytecodeCacheSkipsCacheFile(t *testing.T) {
	withScratchCacheDir(t)

	DisableBytecodeCache = true
	t.Cleanup(func() { DisableBytecodeCache = false })

	dir := t.TempDir()
	path := filepath.Join(dir, "script.lua")
	if err := os.WriteFile(path, []byte("return 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewRuntime()
	if err := r.DoFile(path); err != nil {
		t.Fatalf("DoFile: %v", err)
	}
	cachePath, err := cacheFilePath(path)
	if err != nil {
		t.Fatalf("cacheFileFor: %v", err)
	}
	if _, err := os.Stat(cachePath); err == nil {
		t.Fatal("expected no cache file when disabled")
	}
}
