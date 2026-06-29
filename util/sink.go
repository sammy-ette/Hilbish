package util

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sammy-ette/hilbish/moonlight"
)

var sinkMetaKey = moonlight.StringValue("hshsink")

// #type
// A sink is a structure that has input and/or output to/from a desination.
type Sink struct {
	Rw        *bufio.ReadWriter
	file      *os.File
	UserData  *moonlight.UserData
	autoFlush bool
}

func SinkLoader(mlr *moonlight.Runtime) *moonlight.Table {
	sinkMethods := moonlight.NewTable()
	sinkFuncs := map[string]moonlight.Export{
		"flush":     {Function: luaSinkFlush, ArgNum: 1, Variadic: false},
		"read":      {Function: luaSinkRead, ArgNum: 1, Variadic: false},
		"readAll":   {Function: luaSinkReadAll, ArgNum: 1, Variadic: false},
		"autoFlush": {Function: luaSinkAutoFlush, ArgNum: 2, Variadic: false},
		"write":     {Function: luaSinkWrite, ArgNum: 2, Variadic: false},
		"writeln":   {Function: luaSinkWriteln, ArgNum: 2, Variadic: false},
	}
	mlr.SetExports(sinkMethods, sinkFuncs)

	sinkMeta := moonlight.NewTable()
	sinkIndex := func(mlr *moonlight.Runtime) error {
		s, _ := sinkArg(mlr, 0)

		arg := mlr.Arg(1)
		val := sinkMethods.Get(arg)

		if val != moonlight.NilValue {
			mlr.PushNext1(val)
			return nil
		}

		keyStr, _ := arg.TryString()

		switch keyStr {
		case "pipe":
			val = moonlight.BoolValue(false)
			if s.file != nil {
				fileInfo, _ := s.file.Stat()
				val = moonlight.BoolValue(fileInfo.Mode()&os.ModeCharDevice == 0)
			}
		}

		mlr.PushNext(val)
		return nil
	}

	sinkMeta.Set(moonlight.StringValue("__index"), moonlight.FunctionValue(moonlight.NewGoFunction(mlr, sinkIndex, "__index", 2, false)))
	mlr.SetRegistry(sinkMetaKey, moonlight.TableValue(sinkMeta))

	exports := map[string]moonlight.Export{
		"new": {Function: luaSinkNew, ArgNum: 0, Variadic: false},
	}

	mod := moonlight.NewTable()
	mlr.SetExports(mod, exports)

	SetField(mod, "stderr", moonlight.UserDataValue(NewSink(mlr, os.Stderr).UserData))
	SetField(mod, "stdout", moonlight.UserDataValue(NewSink(mlr, os.Stdout).UserData))
	SetField(mod, "stdin", moonlight.UserDataValue(NewSink(mlr, os.Stdin).UserData))

	return mod
}

func luaSinkNew(mlr *moonlight.Runtime) error {
	snk := NewSink(mlr, new(bytes.Buffer))

	mlr.PushNext1(moonlight.UserDataValue(snk.UserData))
	return nil
}

// #member
// readAll() -> string
// --- @returns string
// Reads all input from the sink.
func luaSinkReadAll(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}

	s, err := sinkArg(mlr, 0)
	if err != nil {
		return err
	}

	if s.autoFlush {
		s.Rw.Flush()
	}

	lines := []string{}
	for {
		line, err := s.Rw.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// We still want to add the data we read
				lines = append(lines, line)
				break
			}

			return err
		}

		lines = append(lines, line)
	}

	mlr.PushNext1(moonlight.StringValue(strings.Join(lines, "")))
	return nil
}

// #member
// read() -> string
// --- @returns string
// Reads a liine of input from the sink.
func luaSinkRead(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}

	s, err := sinkArg(mlr, 0)
	if err != nil {
		return err
	}

	str, _ := s.Rw.ReadString('\n')
	mlr.PushNext(moonlight.StringValue(str))

	return nil
}

// #member
// write(str)
// Writes data to a sink.
func luaSinkWrite(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}

	s, err := sinkArg(mlr, 0)
	if err != nil {
		return err
	}
	data, err := mlr.StringArg(1)
	if err != nil {
		return err
	}

	s.Rw.Write([]byte(data))
	if s.autoFlush {
		s.Rw.Flush()
	}

	return nil
}

// #member
// writeln(str)
// Writes data to a sink with a newline at the end.
func luaSinkWriteln(mlr *moonlight.Runtime) error {
	if err := mlr.CheckNArgs(2); err != nil {
		return err
	}

	s, err := sinkArg(mlr, 0)
	if err != nil {
		return err
	}
	data, err := mlr.StringArg(1)
	if err != nil {
		return err
	}

	s.Rw.Write([]byte(data + "\n"))
	if s.autoFlush {
		s.Rw.Flush()
	}

	return nil
}

// #member
// flush()
// Flush writes all buffered input to the sink.
func luaSinkFlush(mlr *moonlight.Runtime) error {
	if err := mlr.Check1Arg(); err != nil {
		return err
	}

	s, err := sinkArg(mlr, 0)
	if err != nil {
		return err
	}

	s.Rw.Flush()

	return nil
}

// #member
// autoFlush(auto)
// Sets/toggles the option of automatically flushing output.
// A call with no argument will toggle the value.
// --- @param auto boolean|nil
func luaSinkAutoFlush(mlr *moonlight.Runtime) error {
	s, err := sinkArg(mlr, 0)
	if err != nil {
		return err
	}

	v := mlr.Arg(1)
	if v.Type() != moonlight.BoolType && v.Type() != moonlight.NilType {
		return fmt.Errorf("#1 must be a boolean")
	}

	value := !s.autoFlush
	if v.Type() == moonlight.BoolType {
		value = v.AsBool()
	}

	s.autoFlush = value

	return nil
}

func NewSink(mlr *moonlight.Runtime, Rw io.ReadWriter) *Sink {
	s := &Sink{
		Rw:        bufio.NewReadWriter(bufio.NewReader(Rw), bufio.NewWriter(Rw)),
		autoFlush: true,
	}
	s.UserData = sinkUserData(mlr, s)

	if f, ok := Rw.(*os.File); ok {
		s.file = f
	}

	return s
}

func NewSinkInput(mlr *moonlight.Runtime, r io.Reader) *Sink {
	s := &Sink{
		Rw: bufio.NewReadWriter(bufio.NewReader(r), nil),
	}
	s.UserData = sinkUserData(mlr, s)

	if f, ok := r.(*os.File); ok {
		s.file = f
	}

	return s
}

func NewSinkOutput(mlr *moonlight.Runtime, w io.Writer) *Sink {
	s := &Sink{
		Rw:        bufio.NewReadWriter(nil, bufio.NewWriter(w)),
		autoFlush: true,
	}
	s.UserData = sinkUserData(mlr, s)

	if f, ok := w.(*os.File); ok {
		s.file = f
	}

	return s
}

func sinkArg(mlr *moonlight.Runtime, arg int) (*Sink, error) {
	s, ok := valueToSink(mlr.Arg(arg))
	if !ok {
		return nil, fmt.Errorf("#%d must be a sink", arg+1)
	}

	return s, nil
}

func valueToSink(val moonlight.Value) (*Sink, bool) {
	u, ok := val.TryUserData()
	if !ok {
		return nil, false
	}

	s, ok := u.Value().(*Sink)
	return s, ok
}

func sinkUserData(mlr *moonlight.Runtime, s *Sink) *moonlight.UserData {
	sinkMeta := mlr.Registry(sinkMetaKey)
	return moonlight.NewUserData(s, moonlight.ToTable(sinkMeta))
}
