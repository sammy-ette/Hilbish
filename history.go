package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	rt "github.com/arnodel/golua/runtime"
)

type luaHistory struct {}

func (h *luaHistory) Write(line string) (int, error) {
	histWrite := hshMod.Get(rt.StringValue("history")).AsTable().Get(rt.StringValue("add"))
	ln, err := rt.Call1(l.MainThread(), histWrite, rt.StringValue(line))

	var num int64
	if ln.Type() == rt.IntType {
		num = ln.AsInt()
	}

	return int(num), err
}

func (h *luaHistory) GetLine(idx int) (string, error) {
	histGet := hshMod.Get(rt.StringValue("history")).AsTable().Get(rt.StringValue("get"))
	lcmd, err := rt.Call1(l.MainThread(), histGet, rt.IntValue(int64(idx)))

	var cmd string
	if lcmd.Type() == rt.StringType {
		cmd = lcmd.AsString()
	}

	return cmd, err
}

func (h *luaHistory) Len() int {
	histSize := hshMod.Get(rt.StringValue("history")).AsTable().Get(rt.StringValue("size"))
	ln, _ := rt.Call1(l.MainThread(), histSize)

	var num int64
	if ln.Type() == rt.IntType {
		num = ln.AsInt()
	}

	return int(num)
}

func (h *luaHistory) Dump() interface{} {
	// hilbish.history interface already has all function, this isnt used in readline
	return nil
}

// encodeHistoryLine escapes backslashes and newlines for newline-delimited storage.
// Allows multiline history entries to be stored correctly.
func encodeHistoryLine(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	return s
}

// decodeHistoryLine reverses encodeHistoryLine escaping.
// Handles trailing lone backslash edge case.
func decodeHistoryLine(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			next := s[i+1]
			if next == '\\' {
				result.WriteByte('\\')
				i++ // skip the next backslash
			} else if next == 'n' {
				result.WriteByte('\n')
				i++ // skip the 'n'
			} else {
				result.WriteByte(s[i])
			}
		} else {
			result.WriteByte(s[i])
		}
	}
	return result.String()
}

type fileHistory struct {
	items []string
	f *os.File
}

func newFileHistory(path string) *fileHistory {
	dir := filepath.Dir(path)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		panic(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			panic(err)
		}
	}

	lines := strings.Split(string(data), "\n")
	itms := make([]string, len(lines) - 1)
	for i, l := range lines {
		if i == len(lines) - 1 {
			continue
		}
		itms[i] = decodeHistoryLine(l)
	}
	f, err := os.OpenFile(path, os.O_APPEND | os.O_WRONLY | os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}

	fh := &fileHistory{
		items: itms,
		f: f,
	}

	return fh
}

func (h *fileHistory) Write(line string) (int, error) {
	if line == "" {
		return len(h.items), nil
	}

	encodedLine := encodeHistoryLine(line)
	_, err := h.f.WriteString(encodedLine + "\n")
	if err != nil {
		return 0, err
	}
	h.f.Sync()

	h.items = append(h.items, line)
	return len(h.items), nil
}

func (h *fileHistory) GetLine(idx int) (string, error) {
	if len(h.items) == 0 {
		return "", nil
	}
	if idx == -1 { // this should be fixed readline side
		return "", nil
	}
	return h.items[idx], nil
}

func (h *fileHistory) Len() int {
	return len(h.items)
}

func (h *fileHistory) Dump() interface{} {
	return h.items
}

func (h *fileHistory) clear() {
	h.items = []string{}
	h.f.Truncate(0)
	h.f.Sync()
}
