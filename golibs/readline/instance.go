package readline

import (
	"bufio"
	"os"
	"regexp"
	"sync"

	"github.com/arnodel/golua/lib/packagelib"
)

// Instance is used to encapsulate the parameter group and run time of any given
// readline instance so that you can reuse the readline API for multiple entry
// captures without having to repeatedly unload configuration.

// #type
type Readline struct {

	//
	// Input Modes  -------------------------------------------------------------------------------

	// InputMode - The shell can be used in Vim editing mode, or Emacs (classic).
	InputMode InputMode

	// Vim parameters/functions
	// ShowVimMode - If set to true, a string '[i]' or '[N]' indicating the
	// current Vim mode will be appended to the prompt variable, therefore added to
	// the user's custom prompt is set. Applies for both single and multiline prompts
	ShowVimMode     bool
	VimModeColorize bool // If set to true, varies colors of the VimModePrompt

	//
	// Prompt -------------------------------------------------------------------------------------

	Multiline       bool   // If set to true, the shell will have a two-line prompt.
	MultilinePrompt string // If multiline is true, this is the content of the 2nd line.

	mainPrompt     string // If multiline true, the full prompt string / If false, the 1st line of the prompt
	rightPrompt    string
	rightPromptLen int
	realPrompt     []rune // The prompt that is actually on the same line as the beginning of the input line.
	defaultPrompt  []rune
	promptLen      int
	stillOnRefresh bool // True if some logs have printed asynchronously since last loop. Check refresh prompt funcs

	//
	// Input Line ---------------------------------------------------------------------------------

	// PasswordMask is what character to hide password entry behind.
	// Once enabled, set to 0 (zero) to disable the mask again.
	PasswordMask rune

	// readline operating parameters
	line  []rune // This is the input line, with entered text: full line = mlnPrompt + line
	pos   int
	posX  int // Cursor position X
	fullX int // X coordinate of the full input line, including the prompt if needed.
	posY  int // Cursor position Y (if multiple lines span)
	fullY int // Y offset to the end of input line.

	// Buffer received from host programms
	multiline     []byte
	multisplit    []string
	skipStdinRead bool

	// SyntaxHighlight is a helper function to provide syntax highlighting.
	// Once enabled, set to nil to disable again.
	SyntaxHighlighter func([]rune) string

	//
	// Completion ---------------------------------------------------------------------------------

	// TabCompleter is a simple function that offers completion suggestions.
	// It takes the readline line ([]rune) and cursor pos.
	// Returns a prefix string, and several completion groups with their items and description
	// Asynchronously add/refresh completions
	TabCompleter      func([]rune, int, DelayedTabContext) (string, []*CompletionGroup)
	delayedTabContext DelayedTabContext

	// SyntaxCompletion is used to autocomplete code syntax (like braces and
	// quotation marks). If you want to complete words or phrases then you might
	// be better off using the TabCompletion function.
	// SyntaxCompletion takes the line ([]rune) and cursor position, and returns
	// the new line and cursor position.
	SyntaxCompleter func([]rune, int) ([]rune, int)

	// Asynchronously highlight/process the input line
	DelayedSyntaxWorker func([]rune) []rune
	delayedSyntaxCount  int64

	// MaxTabCompletionRows is the maximum number of rows to display in the tab
	// completion grid.
	MaxTabCompleterRows int // = 4

	// tab completion operating parameters
	tcGroups []*CompletionGroup // All of our suggestions tree is in here
	tcPrefix string             // The current tab completion prefix  aggainst which to build candidates

	modeTabCompletion    bool
	compConfirmWait      bool // When too many completions, we ask the user to confirm with another Tab keypress.
	tabCompletionSelect  bool // We may have completions printed, but no selected candidate yet
	tabCompletionReverse bool // Groups sometimes use this indicator to know how they should handle their index
	tcUsedY              int  // Comprehensive offset of the currently built completions
	completionOpen       bool

	// Candidate /  virtual completion string / etc
	currentComp  []rune // The currently selected item, not yet a real part of the input line.
	lineComp     []rune // Same as rl.line, but with the currentComp inserted.
	lineRemain   []rune // When we complete in the middle of a line, we cut and keep the remain.
	compAddSpace bool   // If true, any candidate inserted into the real line is done with an added space.

	//
	// Completion Search  (Normal & History) -----------------------------------------------------

	modeTabFind  bool           // This does not change, because we will search in all options, no matter the group
	tfLine       []rune         // The current search pattern entered
	modeAutoFind bool           // for when invoked via ^R or ^F outside of [tab]
	searchMode   FindMode       // Used for varying hints, and underlying functions called
	regexSearch  *regexp.Regexp // Holds the current search regex match
	search       string
	mainHist     bool   // Which history stdin do we want
	histInfo     []rune // We store a piece of hist info, for dual history sources
	Searcher     func(string, []string) []string

	//
	// History -----------------------------------------------------------------------------------

	// mainHistory - current mapped to CtrlR by default, with rl.SetHistoryCtrlR()
	mainHistory  History
	mainHistName string
	// altHistory is an alternative history input, in case a console user would
	// like to have two different history flows. Mapped to CtrlE by default, with rl.SetHistoryCtrlE()
	altHistory  History
	altHistName string

	// HistoryAutoWrite defines whether items automatically get written to
	// history.
	// Enabled by default. Set to false to disable.
	HistoryAutoWrite bool // = true

	// history operating params
	lineBuf    string
	histPos    int
	histOffset int
	histNavIdx int // Used for quick history navigation.

	//
	// Info -------------------------------------------------------------------------------------

	// InfoText is a helper function which displays infio text below the prompt.
	// InfoText takes the line input from the prompt and the cursor position.
	// It returns the info text to display.
	InfoText func([]rune, int) []rune

	// InfoColor is any ANSI escape codes you wish to use for info formatting. By
	// default this will just be blue.
	InfoFormatting string

	infoText []rune // The actual info text
	infoY    int    // Offset to info, if it spans multiple lines

	//
	// Hints -----------------------------------------------------------------------------------

	// HintText is a helper function which displays hint text right after the user's input.
	// It takes the line input and cursor position.
	// It returns the hint text to display.
	HintText func([]rune, int) []rune

	// HintFormatting is just a string to use as the formatting for the hint. By default
	// this will be a grey color.
	HintFormatting string

	hintText []rune

	//
	// Vim Operatng Parameters -------------------------------------------------------------------

	modeViMode       ViMode //= vimInsert
	viIteration      string
	viUndoHistory    []undoItem
	viUndoSkipAppend bool
	viIsYanking      bool
	registers        *registers // All memory text registers, can be consulted with Alt"

	//
	// Other -------------------------------------------------------------------------------------

	// TempDirectory is the path to write temporary files when editing a line in
	// $EDITOR. This will default to os.TempDir()
	TempDirectory string

	// GetMultiLine is a callback to your host program. Since multiline support
	// is handled by the application rather than readline itself, this callback
	// is required when calling $EDITOR. However if this function is not set
	// then readline will just use the current line.
	GetMultiLine func([]rune) []rune

	EnableGetCursorPos bool

	// event
	evtKeyPress map[string]func(string, []rune, int) *EventReturn

	// concurency
	mutex sync.Mutex

	ViModeCallback   func(ViMode)
	ViActionCallback func(ViAction, []string)

	RawInputCallback func([]rune) // called on all input

	bufferedOut *bufio.Writer

	Loader packagelib.Loader
}

// NewInstance is used to create a readline instance and initialise it with sane defaults.
func NewInstance() *Readline {
	rl := new(Readline)

	// Prompt
	rl.Multiline = false
	rl.mainPrompt = ""
	rl.defaultPrompt = []rune{}
	rl.promptLen = len(rl.computePrompt())

	// Input Editing
	rl.InputMode = Emacs
	rl.ShowVimMode = true // In case the user sets input mode to Vim, everything is ready.

	// Completion
	rl.MaxTabCompleterRows = 50

	// History
	rl.mainHistory = new(ExampleHistory) // In-memory history by default.
	rl.HistoryAutoWrite = true

	// Others
	rl.InfoFormatting = seqFgBlue
	rl.HintFormatting = "\x1b[2m"
	rl.evtKeyPress = make(map[string]func(string, []rune, int) *EventReturn)
	rl.TempDirectory = os.TempDir()
	rl.Searcher = func(needle string, haystack []string) []string {
		suggs := make([]string, 0)

		var err error
		rl.regexSearch, err = regexp.Compile("(?i)" + string(rl.tfLine))
		if err != nil {
			//rl.RefreshPromptLog(err.Error())
			rl.infoText = []rune(Red("Failed to match search regexp"))
		}

		for _, hay := range haystack {
			if rl.regexSearch == nil {
				continue
			}
			if rl.regexSearch.MatchString(hay) {
				suggs = append(suggs, hay)
			}
		}

		return suggs
	}

	rl.bufferedOut = bufio.NewWriter(os.Stdout)

	rl.Loader = packagelib.Loader{
		Name: "readline",
		Load: rl.luaLoader,
	}

	// Registers
	rl.initRegisters()

	return rl
}
