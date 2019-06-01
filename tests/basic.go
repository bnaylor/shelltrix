package main

import (
	"fmt"

	"github.com/bnaylor/shelltrix"
	"github.com/c-bata/go-prompt"
)

func hndHonk(args []string) error {
	fmt.Printf("HONK>>> ")
	fmt.Println(args)
	return nil
}

func honkSubSuggester(cmdline string) *[]prompt.Suggest {
	return &[]prompt.Suggest{
		{
			Text:        "foo",
			Description: "A foo and his mone are soon parte",
		},
		{
			Text:        "bar",
			Description: "qux baz",
		},
	}
}

func honkHelp(args []string) *string {
	text :=
		`To show off its skills, this German shepherd (German for 'hump')
will lick and lick back on demand. A bit like an enthusiastic monkey
a flock of this dachshund will gobble the meat up with every bite.

But what the hell is a dachshund?`

	return &text
}

func hndLonger(args []string) error {
	fmt.Println("That was something else.")
	return nil
}

func main() {
	shelltrix.CommandAdd(shelltrix.Command{
		Name:        "honk",
		Description: "The sound a goose makes",
		Handler:     hndHonk,
		Secondary:   honkSubSuggester,
		Aliases:     []string{"blart"},
		ExtraHelp:   honkHelp,
	})
	shelltrix.CommandAdd(shelltrix.Command{
		Name:        "longercmd",
		Description: "This is to test help formatting",
		Handler:     hndLonger,
	})

	shelltrix.RunShell()
}
