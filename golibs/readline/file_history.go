package readline

import (
	"errors"
	"hilbish/moonlight"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type fileHistory struct {
	items []string
	f     *os.File
}

func newFileHistory(path string) *fileHistory {
	dir := filepath.Dir(path)

	err := os.MkdirAll(dir, 0o644)
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
	itms := make([]string, len(lines)-1)
	for i, l := range lines {
		if i == len(lines)-1 {
			continue
		}
		itms[i] = l
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}

	return &fileHistory{
		items: itms,
		f:     f,
	}
}

func (h *fileHistory) Write(line string) (int, error) {
	if line == "" {
		return len(h.items), nil
	}

	_, err := h.f.WriteString(line + "\n")
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
	if idx == -1 {
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

// Delete removes the history entry at index and rewrites the backing file.
func (h *fileHistory) Delete(index int) error {
	if index < 0 || index >= len(h.items) {
		return nil
	}
	h.items = append(h.items[:index:index], h.items[index+1:]...)
	return h.rewrite()
}

// rewrite truncates the backing file and writes all current items.
func (h *fileHistory) rewrite() error {
	if err := h.f.Truncate(0); err != nil {
		return err
	}
	if _, err := h.f.Seek(0, 0); err != nil {
		return err
	}
	for _, item := range h.items {
		if _, err := h.f.WriteString(item + "\n"); err != nil {
			return err
		}
	}
	return h.f.Sync()
}

// luaHistoryWrapper wraps any Lua table with add/get/size/clear/all methods
// as a readline History interface. This lets users supply custom history handlers.
type luaHistoryWrapper struct {
	handler moonlight.Value
	mlr     *moonlight.Runtime
}

func (h *luaHistoryWrapper) Write(line string) (int, error) {
	addFn := h.handler.AsTable().Get(moonlight.StringValue("add"))
	ln, err := h.mlr.Call1(addFn, moonlight.StringValue(line))
	var num int64
	if ln.Type() == moonlight.IntType {
		num = ln.AsInt()
	}
	return int(num), err
}

func (h *luaHistoryWrapper) GetLine(idx int) (string, error) {
	getFn := h.handler.AsTable().Get(moonlight.StringValue("get"))
	lcmd, err := h.mlr.Call1(getFn, moonlight.IntValue(int64(idx)))
	var cmd string
	if lcmd.Type() == moonlight.StringType {
		cmd = lcmd.AsString()
	}
	return cmd, err
}

func (h *luaHistoryWrapper) Len() int {
	sizeFn := h.handler.AsTable().Get(moonlight.StringValue("size"))
	ln, _ := h.mlr.Call1(sizeFn)
	var num int64
	if ln.Type() == moonlight.IntType {
		num = ln.AsInt()
	}
	return int(num)
}

func (h *luaHistoryWrapper) Dump() interface{} {
	return nil
}

// Delete implements DeletableHistory for a Lua-backed history that exposes a
// "delete" function in its handler table.
func (h *luaHistoryWrapper) Delete(index int) error {
	deleteFn := h.handler.AsTable().Get(moonlight.StringValue("delete"))
	if deleteFn == moonlight.NilValue {
		return nil
	}
	_, err := h.mlr.Call1(deleteFn, moonlight.IntValue(int64(index)))
	return err
}
