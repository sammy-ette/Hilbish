package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"hilbish/golibs/bait"
	"hilbish/golibs/commander"
	"hilbish/util"

	rt "github.com/arnodel/golua/runtime"
	"github.com/pborman/getopt"
	"golang.org/x/term"
)

var (
	l  *rt.Runtime
	lr *lineReader

	luaCompletions = map[string]*rt.Closure{}

	confDir     string
	userDataDir string
	curuser     *user.User

	hooks           *bait.Bait
	cmds            *commander.Commander
	confPath        string
	defaultHistPath string
)

func main() {
	if runtime.GOOS == "linux" {
		// dataDir should only be empty on linux to allow XDG_DATA_DIRS searching.
		// but since it might be set on some distros (nixos) we should still check if its really is empty.
		if dataDir == "" {
			searchableDirs := getenv("XDG_DATA_DIRS", "/usr/local/share/:/usr/share/")
			dataDir = "."
			for _, path := range strings.Split(searchableDirs, ":") {
				_, err := os.Stat(filepath.Join(path, "hilbish", ".hilbishrc.lua"))
				if err == nil {
					dataDir = filepath.Join(path, "hilbish")
					break
				}
			}
		}
	}

	curuser, _ = user.Current()
	confDir, _ = os.UserConfigDir()

	// i honestly dont know what directories to use for this
	switch runtime.GOOS {
	case "linux", "darwin":
		userDataDir = getenv("XDG_DATA_HOME", curuser.HomeDir+"/.local/share")
	default:
		// this is fine on windows, dont know about others
		userDataDir = confDir
	}

	if defaultConfDir == "" {
		// we'll add *our* default if its empty (wont be if its changed comptime)
		defaultConfDir = filepath.Join(confDir, "hilbish")
	} else {
		// else do ~ substitution
		defaultConfDir = filepath.Join(util.ExpandHome(defaultConfDir), "hilbish")
	}
	defaultConfPath := filepath.Join(defaultConfDir, "init.lua")
	if defaultHistDir == "" {
		defaultHistDir = filepath.Join(userDataDir, "hilbish")
	} else {
		defaultHistDir = filepath.Join(util.ExpandHome(defaultHistDir), "hilbish")
	}
	defaultHistPath = filepath.Join(defaultHistDir, ".hilbish-history")
	helpflag := getopt.BoolLong("help", 'h', "Prints Hilbish flags")
	verflag := getopt.BoolLong("version", 'v', "Prints Hilbish version")
	setshflag := getopt.BoolLong("setshellenv", 'S', "Sets $SHELL to Hilbish's executed path")
	cmdflag := getopt.StringLong("command", 'c', "", "Executes a command on startup")
	configflag := getopt.StringLong("config", 'C', defaultConfPath, "Sets the path to Hilbish's config")
	getopt.BoolLong("login", 'l', "Force Hilbish to be a login shell")
	getopt.BoolLong("interactive", 'i', "Force Hilbish to be an interactive shell")
	getopt.BoolLong("noexec", 'n', "Don't execute and only report Lua syntax errors")

	getopt.Parse()
	loginshflag := getopt.Lookup('l').Seen()
	interactiveflag := getopt.Lookup('i').Seen()
	noexecflag := getopt.Lookup('n').Seen()
	confPath = *configflag

	if *helpflag {
		getopt.PrintUsage(os.Stdout)
		os.Exit(0)
	}

	if *cmdflag == "" || interactiveflag {
		interactive = true
	}

	if fileInfo, _ := os.Stdin.Stat(); (fileInfo.Mode()&os.ModeCharDevice) == 0 || !term.IsTerminal(int(os.Stdin.Fd())) {
		interactive = false
	}

	if getopt.NArgs() > 0 {
		interactive = false
	}

	if noexecflag {
		noexecute = true
	}

	// first arg, first character
	if loginshflag || os.Args[0][0] == '-' {
		login = true
	}

	if *verflag {
		fmt.Printf("Hilbish %s\nCompiled with %s\n", getVersion(), runtime.Version())
		os.Exit(0)
	}

	// Set $SHELL if the user wants to
	if *setshflag {
		os.Setenv("SHELL", "hilbish")

		path, err := exec.LookPath("hilbish")
		if err == nil {
			os.Setenv("SHELL", path)
		}

	}

	lr = newLineReader("", false)
	luaInit()

	go handleSignals()

	if fileInfo, _ := os.Stdin.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(bufio.NewReader(os.Stdin))
		for scanner.Scan() {
			text := scanner.Text()
			runInput(text, true)
		}
		exit(0)
	}

	if *cmdflag != "" {
		runInput(*cmdflag, true)
	}

	if getopt.NArgs() > 0 {
		err := util.DoFile(l, getopt.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			exit(1)
		}
		exit(0)
	}

	initialized = true
	/*
		input:
			for interactive {
				running = false

				input, err := lr.Read()

				if err == io.EOF {
					// Exit if user presses ^D (ctrl + d)
					hooks.Emit("hilbish.exit")
					break
				}
				if err != nil {
					if err == readline.CtrlC {
						fmt.Println("^C")
						hooks.Emit("hilbish.cancel")
					} else {
						// If we get a completely random error, print
						fmt.Fprintln(os.Stderr, err)
						if errors.Is(err, syscall.ENOTTY) {
							// what are we even doing here?
							panic("not a tty")
						}
						<-make(chan struct{})
					}
					continue
				}
				var priv bool
				if strings.HasPrefix(input, " ") {
					priv = true
				}

				input = strings.TrimSpace(input)
				if len(input) == 0 {
					running = true
					hooks.Emit("command.exit", 0)
					continue
				}

				if strings.HasSuffix(input, "\\") {
					print("\n")
					for {
						input, err = continuePrompt(strings.TrimSuffix(input, "\\")+"\n", false)
						if err != nil {
							running = true
							lr.SetPrompt(fmtPrompt(prompt))
							goto input // continue inside nested loop
						}
						if !strings.HasSuffix(input, "\\") {
							break
						}
					}
				}

				runInput(input, priv)

				termwidth, _, err := term.GetSize(0)
				if err != nil {
					continue
				}
				fmt.Printf("\u001b[7m∆\u001b[0m" + strings.Repeat(" ", termwidth-1) + "\r")
			}
	*/
	exit(0)
}

func removeDupes(slice []string) []string {
	all := make(map[string]bool)
	newSlice := []string{}
	for _, item := range slice {
		if _, val := all[item]; !val {
			all[item] = true
			newSlice = append(newSlice, item)
		}
	}

	return newSlice
}

func exit(code int) {
	jobs.stopAll()

	// wait for all timers to finish before exiting.
	// only do that when not interactive
	if !interactive {
		timers.wait()
	}

	os.Exit(code)
}

func getVersion() string {
	v := strings.Builder{}

	v.WriteString(ver)
	if gitBranch != "" && gitBranch != "HEAD" {
		v.WriteString("-" + gitBranch)
	}

	if gitCommit != "" {
		v.WriteString("." + gitCommit)
	}

	v.WriteString(" (" + releaseName + ")")

	return v.String()
}

func cut(slice []string, idx int) []string {
	return append(slice[:idx], slice[idx+1:]...)
}
