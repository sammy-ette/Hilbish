package readline

import "testing"

func TestSingleCompletionIsInsertedWithoutOpeningMenu(t *testing.T) {
	rl := newTestRL("ec")
	group := &CompletionGroup{
		DisplayType: TabDisplayGrid,
		NoSpace:     true,
		Items:       []MenuItem{{Value: "echo"}},
	}
	rl.tcPrefix = "ec"
	rl.tcGroups = []*CompletionGroup{group}
	group.init(rl)
	group.goFirstCell()

	rl.updateVirtualComp()

	if got := string(rl.line); got != "echo" {
		t.Fatalf("line after one-candidate completion = %q, want echo", got)
	}
	if rl.modeTabCompletion {
		t.Fatal("single-candidate completion left the menu open")
	}
}

func TestPromptRefreshDoesNotRefetchOpenCompletion(t *testing.T) {
	rl := newTestRL("")
	group := &CompletionGroup{
		DisplayType: TabDisplayGrid,
		Items:       []MenuItem{{Value: "first"}, {Value: "second"}},
	}
	rl.tcGroups = []*CompletionGroup{group}
	group.init(rl)
	group.goFirstCell()
	group.tcPosX = 2
	rl.modeTabCompletion = true
	rl.completionOpen = true

	called := false
	rl.TabCompleter = func([]rune, int, DelayedTabContext) (string, []*CompletionGroup) {
		called = true
		return "", nil
	}
	rl.updateHelpers()

	if called {
		t.Fatal("prompt refresh refetched an already-open completion")
	}
	if group.tcPosX != 2 {
		t.Fatalf("completion selection column = %d, want 2", group.tcPosX)
	}
}

func TestUnicodeCompletionGridRenders(t *testing.T) {
	rl := newTestRL("")
	group := &CompletionGroup{
		DisplayType: TabDisplayGrid,
		MaxLength:   4,
		Items: []MenuItem{
			{Value: "日本語"},
			{Value: "🙂"},
		},
	}
	rl.tcGroups = []*CompletionGroup{group}
	group.init(rl)
	group.goFirstCell()

	if got := group.writeGrid(rl); got == "" {
		t.Fatal("Unicode completion grid rendered an empty result")
	}
}
