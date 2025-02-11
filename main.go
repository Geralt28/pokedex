package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Geralt28/pokedex/pokecache" // Import pokecache from subfolder
)

var cache = pokecache.NewCache()    // Declare globally - cache from second module to use in checking data
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
		"map": {
			name:        "map",
			description: "Displays first/next location areas",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays previous location areas",
			callback:    commandMapb,
		},
		"explore": {
			name:        "explore",
			description: "Show list of all Pokémon's located in the location area",
			callback:    commandExplore,
		},
	}
}

// Struct for commands
type cliCommand struct {
	name        string
	description string
	callback    func(*config, string) error
}

// Structs for JSON parsing
type LocationAreaResponse struct {
	Count    int            `json:"count"`
	Next     string         `json:"next"`
	Previous string         `json:"previous"`
	Results  []LocationArea `json:"results"`
}

type LocationArea struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Nie wiem czy potrzebny i nie wykorzystac wiekszego struct ale jesli bede potrzebowal to mozna go napelnic wskazujac na poprzednia i kolejna strone
type config struct {
	Next     string
	Previous string
}

// do explore - pozniej structy przeniesc do osobnego pliku
// ********************* Start Explore ********************
type LocationExplore struct {
	ID                   int                   `json:"id"`
	Name                 string                `json:"name"`
	GameIndex            int                   `json:"game_index"`
	EncounterMethodRates []EncounterMethodRate `json:"encounter_method_rates"`
	Location             NamedAPIResource      `json:"location"`
	Names                []Name                `json:"names"`
	PokemonEncounters    []PokemonEncounter    `json:"pokemon_encounters"`
}

type EncounterMethodRate struct {
	EncounterMethod NamedAPIResource          `json:"encounter_method"`
	VersionDetails  []EncounterVersionDetails `json:"version_details"`
}

type EncounterVersionDetails struct {
	Rate    int              `json:"rate"`
	Version NamedAPIResource `json:"version"`
}

type NamedAPIResource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Name struct {
	Name     string           `json:"name"`
	Language NamedAPIResource `json:"language"`
}

type PokemonEncounter struct {
	Pokemon        NamedAPIResource         `json:"pokemon"`
	VersionDetails []VersionEncounterDetail `json:"version_details"`
}

type VersionEncounterDetail struct {
	Version          NamedAPIResource  `json:"version"`
	MaxChance        int               `json:"max_chance"`
	EncounterDetails []EncounterDetail `json:"encounter_details"`
}

type EncounterDetail struct {
	MinLevel        int                `json:"min_level"`
	MaxLevel        int                `json:"max_level"`
	ConditionValues []NamedAPIResource `json:"condition_values"`
	Chance          int                `json:"chance"`
	Method          NamedAPIResource   `json:"method"`
}

// ********************* End Explore ********************

func cleanInput(text string) []string {
	return strings.Fields(strings.ToLower(text))
}

func commandMap(cfg *config, _ string) error {
	url := "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20"
	if cfg.Next != "" && cfg.Next != "null" {
		url = cfg.Next
	}

	data, err := czytajAreas(url)
	if err != nil {
		return err
	}

	// Save next and previous pages in config
	cfg.Next = data.Next
	cfg.Previous = data.Previous

	for _, loc := range data.Results {
		fmt.Println(loc.Name)
	}

	return nil
}

func commandMapb(cfg *config, _ string) error {
	url := "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20"
	if cfg.Previous != "" && cfg.Previous != "null" {
		url = cfg.Previous
	} else {
		fmt.Println("you’re on the first page")
		return nil
	}

	data, err := czytajAreas(url)
	if err != nil {
		return err
	}

	// Save next and previous pages in config
	cfg.Next = data.Next
	cfg.Previous = data.Previous

	for _, loc := range data.Results {
		fmt.Println(loc.Name)
	}
	return nil
}

func commandExplore(_ *config, location string) error {
	// zrobic czytanie pola i przez byte wrzucanie do struct. W srodku sprawdzanie cachu lub jesli nie dodawanie do cachu. Nastepnie drukowanie wynikow
	if len(location) == 0 {
		return fmt.Errorf("Brak wskazanego obszaru po explore!")
	}

	url := "https://pokeapi.co/api/v2/location-area/" + location

	data, err := czytajExplore(url)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error: Probably bad area name: %s!", location))
		return err
	}

	//i tu zrobic cos z data - np. wydrukowac, na wzor ponizej, ale zmienic
	fmt.Println(fmt.Sprintf(`Exploring %s...
Found Pokemon:`, location))
	for _, pokemon := range data.PokemonEncounters {
		//for _, loc := range data.Results {
		fmt.Println("-", pokemon.Pokemon.Name)
	}
	return nil
}

func commandExit(cfg *config, _ string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *config, _ string) error {
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

// zamienilem metode odczytu na taka przechodzaca przez bajty zeby w taki sposob cachowac zgodnie z poleceniem
func czytajAreas(url string) (LocationAreaResponse, error) {

	var data LocationAreaResponse

	byte_data, ok := cache.Get(url)
	//if ok {
	//	fmt.Printf("Using cache!") // Debug message
	//} else {
	//	fmt.Println("Cache miss! Fetching from API.") // Debug message
	//}

	if !ok {
		res, err := http.Get(url)
		if err != nil {
			return data, err
		}

		defer res.Body.Close()

		byte_data, err = io.ReadAll(res.Body)
		//err = json.NewDecoder(res.Body).Decode(&data)
		if err != nil {
			return data, err
		}

		cache.Add(url, byte_data)
	}

	if err := json.Unmarshal(byte_data, &data); err != nil {
		return data, err
	}

	return data, nil
}

func czytajExplore(url string) (LocationExplore, error) {

	var data LocationExplore

	byte_data, ok := cache.Get(url)
	//if ok {
	//	fmt.Printf("Using cache!") // Debug message
	//} else {
	//	fmt.Println("Cache miss! Fetching from API.") // Debug message
	//}

	if !ok {
		res, err := http.Get(url)
		if err != nil {
			return data, err
		}

		defer res.Body.Close()

		byte_data, err = io.ReadAll(res.Body)
		//err = json.NewDecoder(res.Body).Decode(&data)
		if err != nil {
			return data, err
		}

		cache.Add(url, byte_data)
	}

	if err := json.Unmarshal(byte_data, &data); err != nil {
		return data, err
	}

	return data, nil
}

func main() {

	const kolor = "\033[34m" // blue
	const reset = "\033[0m"

	fmt.Println("Hello, World!")

	scanner := bufio.NewScanner(os.Stdin)
	cfg := &config{} // Create a config instance
	//pokecache.NewCache()

	for {
		fmt.Printf("%sPokedex > %s", kolor, reset)
		scanner.Scan()
		input := scanner.Text()
		slowa := cleanInput(input)
		dl := len(slowa)
		if dl > 0 {
			slowo := slowa[0]
			//fmt.Printf("Your command was: %s\n", slowa[0])
			polecenie, ok := polecenia[slowo]
			if ok {
				str := ""
				if dl > 1 {
					str = slowa[1]
				}
				polecenie.callback(cfg, str)
			} else {
				fmt.Println("Unknown command")
			}
		}
	}
}
