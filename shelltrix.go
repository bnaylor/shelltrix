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

type HelpFormatHints struct {
	longestCommand     int
	longestDescription int
}

var (
	// Built-in commands only.  Users add commands with CommandAdd
	commandsBuiltIn = map[string]Command{
		"exit": {
			Name:        "exit",
			Handler:     handleExit,
			Description: "Exit this program",
			Aliases:     []string{"quit", "q"},
		},
	}

	// Cache of commands added by external code
	commandsExt = map[string]Command{}

	// Dynamic list of all commands
	commandsAll = map[string]Command{}

	// Reverse map of alises -> commands
	commandAliases = map[string]Command{}

	// Cache toplevel suggestions
	suggestionsTop []prompt.Suggest

	// Active suggestion list
	suggestions *[]prompt.Suggest

	// Compute & cache some hints for help display
	helpHints HelpFormatHints
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
			// Maybe it's an alias?
			alias := aliasSearch(base)
			if alias != nil {
				newargs := args[2:]
				retry := []string{"help", alias.Name}
				retry = append(retry, newargs...)
				fmt.Printf("(FYI: '%s' is an alias for '%s'.)\n\n", base, alias.Name)
				handleHelp(retry)
			} else {
				fmt.Printf("I dunno about '%s'.\n", base)
			}
		}
	} else {
		var c = commandsAll
		for name, val := range c {
			fmt.Printf("  %*s : %-*s", helpHints.longestCommand, name,
				helpHints.longestDescription, val.Description)
			if val.Aliases != nil {
				fmt.Printf("   | Aliases: ")
				for _, a := range val.Aliases {
					fmt.Printf("%s ", a)
				}
			}
			fmt.Printf("\n")
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
		cur := words[len(words)-1]

		if commandsAll[root].Secondary != nil {
			suggestions = commandsAll[root].Secondary(cur)
		} else {
			acmd := aliasSearch(root)
			if acmd != nil {
				suggestions = acmd.Secondary(cur)
			} else {
				suggestions = &[]prompt.Suggest{}
			}
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
	for name, val := range commandAliases {
		t := prompt.Suggest{Text: name, Description: val.Name}
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

// aliasAdd adds an alias for a Command if it doesn't already exist
func aliasAdd(base Command, alias string) {
	existing, exists := commandAliases[alias]
	if exists {
		fmt.Printf("Alias '%s' defined for command '%s', ", alias, base.Name)
		fmt.Printf("but already exists under '%s'. Ignored.\n", existing.Name)
	} else {
		commandAliases[alias] = base
	}
}

// addDefinedAliases sets up any aliases that might be defined in a Command
func addDefinedAliases(cmd Command) {
	if cmd.Aliases != nil {
		for _, a := range cmd.Aliases {
			aliasAdd(cmd, a)
		}
	}
}

// aliasSearch resolves a potential alias to a command
func aliasSearch(c string) *Command {
	entry, exists := commandAliases[c]
	if exists {
		return &entry
	}
	return nil
}

// initCommands initializes base commands
func initCommands() {
	for name, val := range commandsBuiltIn {
		commandsAll[name] = val
		addDefinedAliases(val)
	}
}

// scanHelp runs through all the defined help strings and caches formatting hints
func scanHelp() {
	for name, val := range commandsAll {
		if len(name) > helpHints.longestCommand {
			helpHints.longestCommand = len(name)
		}
		if len(val.Description) > helpHints.longestDescription {
			helpHints.longestDescription = len(val.Description)
		}
	}
}

// CommandAdd adds a new command to the shell
func CommandAdd(cmd Command) {
	commandsExt[cmd.Name] = cmd
	commandsAll[cmd.Name] = cmd
	addDefinedAliases(cmd)
	reinitSuggestions()
	scanHelp()
}

func initShell() {
	initCommands()
	initSuggestions()
	scanHelp()
}

// RunShell starts the input loop and takes over executions
func RunShell() {

	initShell()

	p := prompt.New(nil, completer,
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlC,
			Fn: func(buf *prompt.Buffer) {
				fmt.Println("interrupted")
				os.Exit(130)
			}}),
	)

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
			} else if cmd == "help" || cmd == "?" {
				// help is special..
				handleHelp(cmdargs)
			} else if cmd != "" {
				aliasedCmd := aliasSearch(cmd)
				if aliasedCmd != nil {
					aliasedCmd.Handler(cmdargs)
				} else {
					fmt.Printf("No such command: '%s'\n", cmd)
				}
			}
		}
	}
}
