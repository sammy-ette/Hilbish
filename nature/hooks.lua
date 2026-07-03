-- @hooks
-- Catalog of every signal Hilbish emits via bait. Parsed by cmd/docgen to
-- generate docs/hooks/*. This is the single source of truth for hook docs.
-- Emit sites in Go/Lua source carry a "-- see nature/hooks.lua" pointer comment.

--- @hook command.exit
--- @group command
--- Thrown after the user's ran command is finished.
--- @var code number The exit code of what was executed.
--- @var cmdStr string The command or code that was executed.

--- @hook command.not-executable
--- @group command
--- Thrown when the user attempts to run a file that is not executable
--- (like a text file, or a binary without +x permission).
--- @var cmdStr string The name of the command.

--- @hook command.not-found
--- @group command
--- Thrown if the command attempted to execute was not found.
--- This can be used to customize the text printed when a command is not found.
--- @var cmdStr string The name of the command.
--- @example
--- local bait = require 'bait'
--- -- Remove any present handlers on `command.not-found`
--- local notFoundHooks = bait.hooks 'command.not-found'
--- for _, hook in ipairs(notFoundHooks) do
--- 	bait.release('command.not-found', hook)
--- end
--- -- then assign custom
--- bait.catch('command.not-found', function(cmd)
--- 	print(string.format('The command "%s" was not found.', cmd))
--- end)
--- @example

--- @hook command.preexec
--- @group command
--- Thrown right before a command is executed.
--- @var input string The raw string the user typed, before alias expansion and argument substitution.
--- @var cmdStr string The command that will be directly executed by the current runner.

--- @hook hilbish.cancel
--- @group hilbish
--- Sent when the user cancels their command input with Ctrl-C.

--- @hook hilbish.cd
--- @group hilbish
--- Sent when the current directory of the shell is changed.
--- Since 3.0, hilbish.cd is thrown when fs.cd is called.
--- @var path string Absolute path of the directory that was changed to.
--- @var oldPath string Absolute path of the directory Hilbish was in before the change.

--- @hook hilbish.exit
--- @group hilbish
--- Sent when Hilbish is going to exit.

--- @hook hilbish.notification
--- @group hilbish
--- Thrown when a [notification](../features/notifications) is sent.
--- @var notification table The notification. See the notifications feature doc for its properties.

--- @hook hilbish.rawInput
--- @group hilbish
--- Thrown when the user has entered text in the prompt and pressed enter, before any processing.
--- @var input string The raw input the user typed.

--- @hook hilbish.vimAction
--- @group hilbish
--- Sent when the user does a "vim action," such as yanking or pasting text.
--- See `doc vim-mode actions` for a list of available actions.
--- @var actionName string The name of the Vim action that was done.
--- @var args table Table of args relating to the Vim action.

--- @hook hilbish.vimMode
--- @group hilbish
--- Sent when the Vim mode of Hilbish is changed (like from insert to normal mode).
--- This can be used to change the prompt and notify based on Vim mode.
--- @var modeName string The mode that has been set. Can be: `insert`, `normal`, `delete` or `replace`.

--- @hook job.add
--- @group job
--- Thrown when a new job is added to the job table.
--- @var job table The job object. See the [jobs API](../../api/hilbish/hilbish.jobs) for its properties.

--- @hook job.done
--- @group job
--- Thrown when a background job exits.
--- @var job table The job object. See the [jobs API](../../api/hilbish/hilbish.jobs) for its properties.

--- @hook job.start
--- @group job
--- Thrown when a new background job starts.
--- @var job table The job object. See the [jobs API](../../api/hilbish/hilbish.jobs) for its properties.

--- @hook signal.resize
--- @group signal
--- Thrown when the terminal is resized.

--- @hook signal.sigint
--- @group signal
--- Thrown when Hilbish receives the SIGINT signal (Ctrl-C).

--- @hook signal.sigusr1
--- @group signal
--- Thrown when SIGUSR1 is sent to Hilbish.

--- @hook signal.sigusr2
--- @group signal
--- Thrown when SIGUSR2 is sent to Hilbish.
