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

func main() {
	shelltrix.CommandAdd(shelltrix.Command{
		Name:        "honk",
		Description: "The sound a goose makes",
		Handler:     hndHonk,
		Secondary:   honkSubSuggester,
	})

	shelltrix.RunShell()
}
