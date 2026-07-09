//go:build !midnight

package moonlight

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	rt "github.com/arnodel/golua/runtime"
)

var DisableBytecodeCache = false
var cacheDirOverride string

var (
	cacheDirOnce = new(sync.Once)
	cacheDirPath string
	cacheDirErr  error
)

var binaryFingerprintOverride int64

var (
	binaryFingerprintOnce = new(sync.Once)
	binaryFingerprint     int64
)

// currentBinaryFingerprint returns the running executable's mtime to invalidate cache entries when Hilbish is rebuilt
func currentBinaryFingerprint() int64 {
	binaryFingerprintOnce.Do(func() {
		if binaryFingerprintOverride != 0 {
			binaryFingerprint = binaryFingerprintOverride
			return
		}
		exe, err := os.Executable()
		if err != nil {
			return
		}
		info, err := os.Stat(exe)
		if err != nil {
			return
		}
		binaryFingerprint = info.ModTime().UnixNano()
	})

	return binaryFingerprint
}

func bytecodeCacheDir() (string, error) {
	cacheDirOnce.Do(func() {
		if cacheDirOverride != "" {
			cacheDirPath = cacheDirOverride
			return
		}
		base, err := os.UserCacheDir()
		if err != nil {
			cacheDirErr = err
			return
		}
		cacheDirPath = filepath.Join(base, "hilbish", "bytecode")
	})

	return cacheDirPath, cacheDirErr
}

func cacheFilePath(path string) (string, error) {
	dir, err := bytecodeCacheDir()
	if err != nil {
		return "", err
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}

	sum := sha256.Sum256([]byte(abs))
	return filepath.Join(dir, hex.EncodeToString(sum[:])+".hshbc"), nil
}

func writeCacheEntry(path string, mtime time.Time, bytecode []byte) {
	if DisableBytecodeCache {
		return
	}

	cachePath, err := cacheFilePath(path)
	if err != nil {
		return
	}

	dir := filepath.Dir(cachePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}

	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, currentBinaryFingerprint())
	binary.Write(&buf, binary.LittleEndian, mtime.UnixNano())
	buf.Write(bytecode)

	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(buf.Bytes()); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return
	}

	os.Rename(tmpName, cachePath)
}

func readCacheEntry(path string, mtime time.Time) ([]byte, bool) {
	if DisableBytecodeCache {
		return nil, false
	}

	cachePath, err := cacheFilePath(path)
	if err != nil {
		return nil, false
	}

	f, err := os.Open(cachePath)
	if err != nil {
		return nil, false
	}
	defer f.Close()

	var fingerprint int64
	if err := binary.Read(f, binary.LittleEndian, &fingerprint); err != nil {
		return nil, false
	}
	if fingerprint != currentBinaryFingerprint() {
		return nil, false
	}

	var storedNanos int64
	if err := binary.Read(f, binary.LittleEndian, &storedNanos); err != nil {
		return nil, false
	}
	if storedNanos != mtime.UnixNano() {
		return nil, false
	}

	bytecode, err := io.ReadAll(f)
	if err != nil || !rt.HasMarshalPrefix(bytecode) {
		return nil, false
	}

	return bytecode, true
}

func loadCachedOrCompile(rtm *rt.Runtime, path string, src []byte, env rt.Value, stripComment bool) (*rt.Closure, error) {
	info, statErr := os.Stat(path)

	if statErr == nil {
		if bytecode, ok := readCacheEntry(path, info.ModTime()); ok {
			clos, err := rtm.LoadFromSourceOrCode(path, bytecode, "b", env, false)
			if err == nil {
				return clos, nil
			}
		}
	}

	clos, err := rtm.LoadFromSourceOrCode(path, src, "bt", env, stripComment)
	if err == nil && statErr == nil {
		dumpBudget := rtm.LinearUnused(10)
		var buf bytes.Buffer
		used, dumpErr := rt.MarshalConst(&buf, rt.CodeValue(rtm.RefactorCodeConsts(clos.Code)), dumpBudget)
		rtm.LinearRequire(10, used)
		if dumpErr == nil {
			writeCacheEntry(path, info.ModTime(), buf.Bytes())
		}
	}

	return clos, err
}

func installBytecodeCacheHooks(r *rt.Runtime) {
	env := r.GlobalEnv()
	r.SetEnvGoFunc(env, "dofile", cachedDoFile, 1, false)
	installCachedLuaSearcher(r)
}

func cachedDoFile(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	env := rt.TableValue(t.GlobalEnv())
	next := c.Next()

	if c.NArgs() == 0 {
		budget := t.LinearUnused(10)
		chunk, err := io.ReadAll(io.LimitReader(os.Stdin, int64(budget)))
		if err != nil {
			return nil, err
		}
		t.LinearRequire(10, uint64(len(chunk)))
		clos, err := t.LoadFromSourceOrCode("stdin", chunk, "bt", env, true)
		if err != nil {
			return nil, err
		}
		return clos.Continuation(t, next), nil
	}

	path, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(string(path))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	budget := t.LinearUnused(10)
	reader := io.Reader(f)
	if budget > 0 {
		reader = io.LimitReader(reader, int64(budget))
	}
	chunk, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	t.LinearRequire(10, uint64(len(chunk)))
	clos, err := loadCachedOrCompile(t.Runtime, string(path), chunk, env, true)
	if err != nil {
		return nil, err
	}
	return clos.Continuation(t, next), nil
}

func installCachedLuaSearcher(r *rt.Runtime) {
	pkg, ok := r.GlobalEnv().Get(rt.StringValue("package")).TryTable()
	if !ok {
		return
	}
	searchers, ok := pkg.Get(rt.StringValue("searchers")).TryTable()
	if !ok {
		return
	}

	searchers.Set(rt.IntValue(2), rt.FunctionValue(rt.NewGoFunction(cachedSearchLua, "searchlua", 1, false)))
}

func cachedSearchLua(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.Check1Arg(); err != nil {
		return nil, err
	}
	name, err := c.StringArg(0)
	if err != nil {
		return nil, err
	}

	pkg, ok := t.GlobalEnv().Get(rt.StringValue("package")).TryTable()
	if !ok {
		return nil, errors.New("package table missing")
	}
	pathTemplate, ok := pkg.Get(rt.StringValue("path")).TryString()
	if !ok {
		return nil, errors.New("package.path must be a string")
	}

	found, templates := searchPathTemplates(string(name), string(pathTemplate))
	next := c.Next()
	if found == "" {
		t.Push1(next, rt.StringValue(strings.Join(templates, "\n")))
		return next, nil
	}

	t.Push1(next, rt.FunctionValue(rt.NewGoFunction(cachedLoadLua, "loadlua", 2, false)))
	t.Push1(next, rt.StringValue(found))
	return next, nil
}

func cachedLoadLua(t *rt.Thread, c *rt.GoCont) (rt.Cont, error) {
	if err := c.CheckNArgs(2); err != nil {
		return nil, err
	}
	// arg 0 is the module name, only the resolved path (arg 1) is needed to load it
	path, err := c.StringArg(1)
	if err != nil {
		return nil, err
	}

	src, err := os.ReadFile(string(path))
	if err != nil {
		return nil, err
	}

	env := rt.TableValue(t.GlobalEnv())
	clos, err := loadCachedOrCompile(t.Runtime, string(path), src, env, true)
	if err != nil {
		return nil, err
	}

	return rt.Continue(t, rt.FunctionValue(clos), c.Next())
}

func searchPathTemplates(name, pathTemplate string) (found string, templates []string) {
	name = strings.ReplaceAll(name, ".", "/")
	for template := range strings.SplitSeq(pathTemplate, ";") {
		candidate := strings.ReplaceAll(template, "?", name)
		templates = append(templates, candidate)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, templates
		}
	}
	return "", templates
}
