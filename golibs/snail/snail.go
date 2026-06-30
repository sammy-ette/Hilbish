package snail

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/sammy-ette/hilbish/moonlight"
	"github.com/sammy-ette/hilbish/util"

	"mvdan.cc/sh/v3/shell"

	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

// #type
// A Snail is a shell script interpreter instance.
type Snail struct {
	runner  *interp.Runner
	runtime *moonlight.Runtime

	// interp.ExecHandlers stacks middleware and our handler never calls next,
	// so only register it once otherwise an old stale one keeps running
	handlerReady     bool
	aliasesListFn    moonlight.Value
	aliasesResolveFn moonlight.Value
	runJobFn         moonlight.Value
	bg               bool
}

func New(mlr *moonlight.Runtime) *Snail {
	runner, _ := interp.New()

	return &Snail{
		runner:  runner,
		runtime: mlr,
	}
}

// Checks if input is incomplete. Does not error otherwise.
func Validate(input string) bool {
	r := strings.NewReader(input)
	_, err := syntax.NewParser().Parse(r, "")
	return !syntax.IsIncomplete(err)
}

func (s *Snail) Run(cmd string, strms *util.Streams) (bool, io.Writer, io.Writer, error) {
	file, err := syntax.NewParser().Parse(strings.NewReader(cmd), "")
	if err != nil {
		return false, nil, nil, err
	}

	if strms == nil {
		strms = &util.Streams{}
	}

	if strms.Stdout == nil {
		strms.Stdout = os.Stdout
	}

	if strms.Stderr == nil {
		strms.Stderr = os.Stderr
	}

	if strms.Stdin == nil {
		strms.Stdin = os.Stdin
	}

	interp.StdIO(strms.Stdin, strms.Stdout, strms.Stderr)(s.runner)
	interp.Env(expand.ListEnviron(append(os.Environ(), "PATH="+os.Getenv("PATH"))...))(s.runner)

	s.ensureHandler()

	var bg bool
	for _, stmt := range file.Stmts {
		bg = stmt.Background
		s.bg = stmt.Background

		err = s.runner.Run(context.TODO(), stmt)
		if err != nil {
			return bg, strms.Stdout, strms.Stderr, err
		}
	}

	return bg, strms.Stdout, strms.Stderr, nil
}

func (s *Snail) ensureHandler() {
	if s.handlerReady {
		return
	}
	s.handlerReady = true

	aliasesMod := s.runtime.MustDoString("return hilbish.aliases").AsTable()
	s.aliasesListFn = aliasesMod.Get(moonlight.StringValue("list"))
	s.aliasesResolveFn = aliasesMod.Get(moonlight.StringValue("resolve"))

	s.runJobFn = s.runtime.MustDoString(`
		return function(cmdstr, opts, background)
			local j = hilbish.jobs.add(cmdstr, opts)
			return j:start{ background = background }
		end
	`)

	execHandler := func(next interp.ExecHandlerFunc) interp.ExecHandlerFunc {
		return func(ctx context.Context, args []string) error {
			argstring := strings.Join(args, " ")
			// i dont really like this but it works
			aliases := make(map[string]string)
			aliasesLua, err := s.runtime.Call1(s.aliasesListFn)
			if err != nil {
				return err
			}
			moonlight.ForEach(moonlight.ToTable(aliasesLua), func(k, v moonlight.Value) {
				aliases[k.AsString()] = v.AsString()
			})
			if aliases[args[0]] != "" {
				for i, arg := range args {
					if strings.Contains(arg, " ") {
						args[i] = fmt.Sprintf("\"%s\"", arg)
					}
				}
				argstring = strings.Join(args, " ")

				// If alias was found, use command alias
				resolved, err := s.runtime.Call1(s.aliasesResolveFn, moonlight.StringValue(argstring))
				if err != nil {
					return err
				}
				argstring = resolved.AsString()

				args, err = shell.Fields(argstring, nil)
				if err != nil {
					return err
				}
			}

			// If command is defined in Lua then run it
			luacmdArgs := moonlight.NewTable()
			for i, str := range args[1:] {
				luacmdArgs.Set(moonlight.IntValue(int64(i+1)), moonlight.StringValue(str))
			}

			hc := interp.HandlerCtx(ctx)

			cmds := make(map[string]*moonlight.Closure)
			luaCmds := moonlight.ToTable(s.runtime.MustDoString("local commander = require 'commander'; return commander.registry()"))
			moonlight.ForEach(luaCmds, func(k, v moonlight.Value) {
				cmds[k.AsString()] = v.AsTable().Get(moonlight.StringValue("exec")).AsClosure()
			})
			if cmd := cmds[args[0]]; cmd != nil {
				stdin := util.NewSinkInput(s.runtime, hc.Stdin)
				stdout := util.NewSinkOutput(s.runtime, hc.Stdout)
				stderr := util.NewSinkOutput(s.runtime, hc.Stderr)

				sinks := moonlight.NewTable()
				sinks.Set(moonlight.StringValue("in"), moonlight.UserDataValue(stdin.UserData))
				sinks.Set(moonlight.StringValue("input"), moonlight.UserDataValue(stdin.UserData))
				sinks.Set(moonlight.StringValue("out"), moonlight.UserDataValue(stdout.UserData))
				sinks.Set(moonlight.StringValue("err"), moonlight.UserDataValue(stderr.UserData))

				t := moonlight.NewThread(s.runtime)
				sig := make(chan os.Signal, 1)
				exit := make(chan bool, 1)
				done := make(chan struct{})

				luaexitcode := moonlight.IntValue(63)
				var err error

				signal.Notify(sig, os.Interrupt)
				go func() {
					select {
					case <-sig:
						t.Kill()
					case <-done: // branch allows the goroutine to go away
					}
				}()

				go func() {
					luaexitcode, err = t.Call1(moonlight.FunctionValue(cmd), moonlight.TableValue(luacmdArgs), moonlight.TableValue(sinks))
					exit <- true
				}()

				<-exit
				close(done)
				signal.Stop(sig)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error in command:\n"+err.Error())
					return interp.NewExitStatus(1)
				}

				var exitcode uint8

				if code, ok := luaexitcode.TryInt(); ok {
					exitcode = uint8(code)
				} else if luaexitcode != moonlight.NilValue {
					commanderMod := s.runtime.MustDoString("return require 'commander'").AsTable()
					deregister := commanderMod.Get(moonlight.StringValue("deregister"))
					s.runtime.Call1(deregister, moonlight.StringValue(args[0]))
					fmt.Fprintf(os.Stderr, "Commander did not return number for exit code. %s, you're fired.\n", args[0])
				}

				return interp.NewExitStatus(exitcode)
			}

			path, err := util.LookPath(args[0])
			if err == util.ErrNotExec {
				return util.ExecError{
					Typ:   "not-executable",
					Cmd:   args[0],
					Code:  126,
					Colon: true,
					Err:   util.ErrNotExec,
				}
			} else if err != nil {
				return util.ExecError{
					Typ:  "not-found",
					Cmd:  args[0],
					Code: 127,
					Err:  util.ErrNotFound,
				}
			}

			env := hc.Env
			envList := os.Environ()
			env.Each(func(name string, vr expand.Variable) bool {
				if name == "PATH" {
					return true
				}
				if vr.Exported && vr.Kind == expand.String {
					envList = append(envList, name+"="+vr.String())
				}
				return true
			})

			optsTbl := moonlight.NewTable()

			argsTbl := moonlight.NewTable()
			for i, a := range args {
				argsTbl.Set(moonlight.IntValue(int64(i+1)), moonlight.StringValue(a))
			}
			optsTbl.Set(moonlight.StringValue("args"), moonlight.TableValue(argsTbl))
			optsTbl.Set(moonlight.StringValue("path"), moonlight.StringValue(path))
			optsTbl.Set(moonlight.StringValue("dir"), moonlight.StringValue(hc.Dir))

			envTbl := moonlight.NewTable()
			for i, e := range envList {
				envTbl.Set(moonlight.IntValue(int64(i+1)), moonlight.StringValue(e))
			}
			optsTbl.Set(moonlight.StringValue("env"), moonlight.TableValue(envTbl))

			sinksTbl := moonlight.NewTable()
			inSink := util.NewSinkInput(s.runtime, hc.Stdin)
			outSink := util.NewSinkOutput(s.runtime, hc.Stdout)
			errSink := util.NewSinkOutput(s.runtime, hc.Stderr)
			sinksTbl.Set(moonlight.StringValue("in"), moonlight.UserDataValue(inSink.UserData))
			sinksTbl.Set(moonlight.StringValue("out"), moonlight.UserDataValue(outSink.UserData))
			sinksTbl.Set(moonlight.StringValue("err"), moonlight.UserDataValue(errSink.UserData))
			optsTbl.Set(moonlight.StringValue("sinks"), moonlight.TableValue(sinksTbl))

			exitVal, err := s.runtime.Call1(s.runJobFn,
				moonlight.StringValue(argstring),
				moonlight.TableValue(optsTbl),
				moonlight.BoolValue(s.bg),
			)

			var exit uint8
			if err != nil {
				exit = util.HandleExecErr(err)
			} else if code, ok := exitVal.TryInt(); ok {
				exit = uint8(code)
			}
			return interp.NewExitStatus(exit)
		}
	}
	interp.ExecHandlers(execHandler)(s.runner)
}
