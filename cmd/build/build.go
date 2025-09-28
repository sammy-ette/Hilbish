package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var PREFIX string
var BINDIR string
var LIBDIR string

func initEnvs(local bool) {
	if pfx := os.Getenv("PREFIX"); pfx != "" {
		PREFIX = pfx
	} else {
		switch runtime.GOOS {
		case "windows":
			// do nothing, idk
		default:
			PREFIX = "/usr/local"
		}
	}

	if ldr := os.Getenv("LIBDIR"); ldr != "" {
		LIBDIR = ldr
	} else {
		switch runtime.GOOS {
		case "windows":
			// is this right?
			LIBDIR = filepath.Join(os.Getenv("APPDATA"), "hilbish")
		default:
			LIBDIR = filepath.Join(PREFIX, "share", "hilbish")
		}
	}

	if bdr := os.Getenv("BINDIR"); bdr != "" {
		BINDIR = bdr
	} else {
		switch runtime.GOOS {
		case "windows":
			// make it the same as libdir, because yes
			BINDIR = LIBDIR
		default:
			BINDIR = filepath.Join(PREFIX, "bin")
		}
	}
}

func main() {
	local := false
	if len(os.Args) >= 3 {
		if os.Args[2] == "local" {
			local = true
		}
	}
	initEnvs(local)

	if len(os.Args) == 1 {
		buildGit()
	} else if os.Args[1] == "stable" {
		buildStable()
	} else if os.Args[1] == "install" {
		install()
	}
}

func buildGit() {
	var libdir string
	switch runtime.GOOS {
	case "linux":
		if ldr := os.Getenv("LIBDIR"); ldr != "" {
			libdir = ldr
		} else {
			libdir = LIBDIR
		}
	default:
		libdir = LIBDIR
	}
	println(libdir)

	commit, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		panic(err)
	}

	branch, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("go", "build", "-ldflags", fmt.Sprintf(`-s -w -X main.dataDir=%s -X main.gitCommit=%s -X main.gitBranch=%s`, libdir, string(commit), string(branch)))
	cmd.Env = env()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	}

	os.Exit(cmd.ProcessState.ExitCode())
}

func buildStable() {
	cmd := exec.Command("go", "build")
	cmd.Env = env()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}

	os.Exit(cmd.ProcessState.ExitCode())
}

func env() []string {
	envs := os.Environ()
	for _, val := range envs {
		if envkv := strings.Split(val, "="); envkv[0] == "OS" {
			envs = append(envs, "GOOS="+envkv[1])
		}

		if envkv := strings.Split(val, "="); envkv[0] == "ARCH" {
			envs = append(envs, "GOARCH="+envkv[1])
		}
	}

	return envs
}

func install() {
	fmt.Println("bindir", BINDIR)
	os.MkdirAll(BINDIR, 0755)

	hilbishPath := filepath.Join(BINDIR, "hilbish")
	err := copyFile("hilbish", hilbishPath)
	if err != nil {
		panic(err)
	}

	fmt.Println("libdir", LIBDIR)
	err = os.MkdirAll(LIBDIR, 0755)
	if err != nil {
		panic(err)
	}

	dirs := []string{"libs", "docs", "emmyLuaDocs", "nature"}
	for _, dir := range dirs {
		err := copyFS(filepath.Join(LIBDIR, dir), os.DirFS(dir))
		if err != nil {
			panic(err)
		}
	}

	err = copyFile(".hilbishrc.lua", filepath.Join(LIBDIR, ".hilbishrc.lua"))
	if err != nil {
		panic(err)
	}

	// i guess checking if not windows makes more sense..
	if runtime.GOOS != "windows" {
		shells, err := os.ReadFile("/etc/shells")
		if err != nil {
			// pass, i guess
			return
		}

		if !bytes.Contains(shells, []byte(hilbishPath)) {
			f, err := os.OpenFile("/etc/shells", os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				panic(err)
			}

			f.Write([]byte(hilbishPath))
			f.Write([]byte("\n"))
		}
	}
}

// https://pkg.go.dev/os#CopyFS that doesnt error if a file already exists
func copyFS(dir string, fsys fs.FS) error {
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		fpath, err := filepath.Localize(path)
		if err != nil {
			return err
		}
		newPath := filepath.Join(dir, fpath)
		if d.IsDir() {
			return os.MkdirAll(newPath, 0777)
		}

		// TODO(panjf2000): handle symlinks with the help of fs.ReadLinkFS
		// 		once https://go.dev/issue/49580 is done.
		//		we also need filepathlite.IsLocal from https://go.dev/cl/564295.
		if !d.Type().IsRegular() {
			return &os.PathError{Op: "CopyFS", Path: path, Err: os.ErrInvalid}
		}

		r, err := fsys.Open(path)
		if err != nil {
			return err
		}
		defer r.Close()
		info, err := r.Stat()
		if err != nil {
			return err
		}
		w, err := os.OpenFile(newPath, os.O_CREATE|os.O_WRONLY, 0666|info.Mode()&0777)
		if err != nil {
			return err
		}

		if _, err := io.Copy(w, r); err != nil {
			w.Close()
			return &os.PathError{Op: "Copy", Path: newPath, Err: err}
		}
		return w.Close()
	})
}

func copyFile(from, to string) error {
	exe, err := os.ReadFile(from)
	if err != nil {
		return err
	}

	return os.WriteFile(to, exe, 0755)
}
