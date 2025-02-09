package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type cliCommand struct {
	name        string
	description string
	callback    func() error
}

var polecenia map[string]cliCommand // Declare globally

func init() {
	polecenia = map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
	}
}

func cleanInput(text string) []string {
	return strings.Fields(strings.ToLower(text))
}

func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp() error {
	tresc := `Welcome to the Pokedex!
Usage:

`
	for _, c := range polecenia {
		dodaj_txt := fmt.Sprintf("%s: %s\n", c.name, c.description)
		tresc = tresc + dodaj_txt
	}
	fmt.Print(tresc)
	return nil
}

func main() {

	fmt.Println("Hello, World!")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		input := scanner.Text()
		slowa := cleanInput(input)
		if len(slowa) > 0 {
			slowo := slowa[0]
			//fmt.Printf("Your command was: %s\n", slowa[0])
			polecenie, ok := polecenia[slowo]
			if ok {
				polecenie.callback()
			} else {
				fmt.Println("Unknown command")
			}
		}
	}
}
