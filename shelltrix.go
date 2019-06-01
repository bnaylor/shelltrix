package shelltrix

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

// ExtraHelpText defines a function for commands to produce more help text
type ExtraHelpText func([]string) *string

// Command maps commands to handlers
//    Required: Name, Handler, Description
//    Optional: Aliases, SecondarySuggester, ExtraHelp
type Command struct {
	Name        string
	Handler     CommandHandler
	Description string
	Aliases     []string
	Secondary   SecondarySuggester
	ExtraHelp   ExtraHelpText
}

var (
	// Built-in commands only.  Users add commands with CommandAdd
	commandsBuiltIn = map[string]Command{
		"exit": {
			Name:        "exit",
			Handler:     handleExit,
			Description: "Exit this program",
			Aliases:     []string{"quit", "bail"},
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
	if len(args) > 1 {
		base := args[1]
		cmd, exists := commandsAll[base]
		if exists {
			h := cmd.ExtraHelp
			if h != nil {
				text := h(args[1:])
				fmt.Println(*text)
			} else {
				fmt.Printf("No more help for '%s'\n", base)
			}
		} else {
			fmt.Printf("I dunno about '%s'.\n", base)
		}
	} else {
		var c = commandsAll
		for name, val := range c {
			fmt.Printf("  %s : %s\n", name, val.Description)
		}
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

// Search through all commands to see if 'c' is an alias to one of them.
// Would obviously be better to build a hash of aliases at init time. TBD
func searchAliases(c string) *Command {
	for _, val := range commandsAll {
		if val.Aliases != nil {
			for _, a := range val.Aliases {
				if a == c {
					return &val
				}
			}
		}
	}
	return nil
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
			} else if cmd == "help" {
				// help is special..
				handleHelp(cmdargs)
			} else if cmd != "" {
				aliasedCmd := searchAliases(cmd)
				if aliasedCmd != nil {
					aliasedCmd.Handler(cmdargs)
				} else {
					fmt.Printf("No such command: '%s'\n", cmd)
				}
			}
		}
	}
}
