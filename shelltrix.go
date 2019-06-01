package shelltrix

// TBD : aliases.  Map 'quit' to exit as metadata instead of a separate command
// TBD : inline completions.  Scoot back in cmdline to correct an argument,
//       tab knows how to complete *that* instead of just what's at the end
//       of the line.

import (
	"fmt"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
)

// CommandHandler defines the function type for handling commands
type CommandHandler func([]string) error

// SecondarySuggester defines a function for replacing suggestions for subcommands
type SecondarySuggester func(string) *[]prompt.Suggest

// Command maps commands to handlers
type Command struct {
	Name        string
	Handler     CommandHandler
	Description string
	Secondary   SecondarySuggester // optional
}

var (
	// Built-in commands only.  Users add commands with CommandAdd
	commandsBuiltIn = map[string]Command{
		"exit": {
			Name:        "exit",
			Handler:     handleExit,
			Description: "Exit this program",
		},
		"quit": {
			Name:        "quit",
			Handler:     handleExit,
			Description: "Exit this program",
		},
	}

	// Cache of commands added by external code
	commandsExt = map[string]Command{}

	// Dynamic list of all commands
	commandsAll = map[string]Command{}

	// Cache toplevel suggestions
	suggestionsTop []prompt.Suggest

	// Active suggestion list
	suggestions *[]prompt.Suggest
)

// =-= HANDLERS =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=
func handleExit(args []string) error {
	os.Exit(0)
	return nil
}

// this can't actually be referenced from the map.  okay, go?
func handleHelp(args []string) error {
	var c = commandsAll
	for name, val := range c {
		fmt.Printf("  %s : %s\n", name, val.Description)
	}
	return nil
}

// =-= </HANDLERS> =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=

func completer(d prompt.Document) []prompt.Suggest {
	if d.Text == "" {
		return nil
	}
	//	fmt.Println(d)
	// If we have something in the commandline and we've just hit 'space',
	// indicating a new 'word', we can replace the suggester with per-command
	// suggestions to allow each command to customize sub-command completions.
	if d.GetWordBeforeCursorWithSpace() != "" && d.GetWordBeforeCursor() == "" {
		// Get the whole commandline as it stands
		cmdline := d.CurrentLine()
		// Pop off the root command to figure out if we have a secondary
		words := strings.Fields(cmdline)
		root := words[0]

		if commandsAll[root].Secondary != nil {
			cur := words[len(words)-1]
			suggestions = commandsAll[root].Secondary(cur)
		} else {
			suggestions = &[]prompt.Suggest{}
		}
	} else if d.CursorPositionCol() == 1 {
		// reset
		suggestions = &suggestionsTop
	}
	return prompt.FilterHasPrefix(*suggestions, d.GetWordBeforeCursor(), true)
}

func initSuggestions() {
	suggestionsTop = []prompt.Suggest{}

	for name, val := range commandsAll {
		t := prompt.Suggest{Text: name, Description: val.Description}
		suggestionsTop = append(suggestionsTop, t)
	}
	suggestionsTop = append(suggestionsTop, prompt.Suggest{Text: "?",
		Description: "Type 'help' to list commands"})
	suggestionsTop = append(suggestionsTop, prompt.Suggest{Text: "help",
		Description: "List commands and what they do"})
}

func reinitSuggestions() {
	initSuggestions()
}

func initCommands() {
	for name, val := range commandsBuiltIn {
		commandsAll[name] = val
	}
}

// CommandAdd adds a new command to the shell
func CommandAdd(cmd Command) {
	commandsExt[cmd.Name] = cmd
	commandsAll[cmd.Name] = cmd
	reinitSuggestions()
}

func initShell() {
	initCommands()
	initSuggestions()
}

// RunShell starts the input loop and takes over executions
func RunShell() {

	initShell()

	p := prompt.New(nil, completer)

	// input loop
	for {
		suggestions = &suggestionsTop
		t := p.Input()
		cmdargs := strings.Fields(t)
		if len(cmdargs) > 0 {
			cmd := cmdargs[0]
			h, exists := commandsAll[cmd]
			if exists {
				h.Handler(cmdargs)
			} else if t == "help" {
				// help is special..
				handleHelp(cmdargs)
			} else if t != "" {
				fmt.Printf("No such command: '%s'\n", cmd)
			}
		}
	}
}
