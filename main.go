package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Geralt28/pokedex/pokecache" // Import pokecache from subfolder
)

const szansaZlapania = 70.0

var cache = pokecache.NewCache()    // Declare globally - cache from second module to use in checking data
var polecenia map[string]cliCommand // Declare globally
var zlapane = make(map[string]Pokemon)

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
		"catch": {
			name:        "catch",
			description: "Try to catch the selected pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspecting caught pokemon",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "List of caught pokemons",
			callback:    commandPokedex,
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

// ******************** Start Pokemon *******************

type Pokemon struct {
	ID                     int                `json:"id"`
	Name                   string             `json:"name"`
	BaseExperience         int                `json:"base_experience"`
	Height                 int                `json:"height"`
	IsDefault              bool               `json:"is_default"`
	Order                  int                `json:"order"`
	Weight                 int                `json:"weight"`
	Abilities              []Ability          `json:"abilities"`
	Forms                  []NamedAPIResource `json:"forms"`
	GameIndices            []GameIndex        `json:"game_indices"`
	HeldItems              []HeldItem         `json:"held_items"`
	LocationAreaEncounters string             `json:"location_area_encounters"`
	Moves                  []Move             `json:"moves"`
	Species                NamedAPIResource   `json:"species"`
	Sprites                Sprites            `json:"sprites"`
	Stats                  []Stat             `json:"stats"`
	Types                  []Type             `json:"types"`
}

type Ability struct {
	IsHidden bool             `json:"is_hidden"`
	Slot     int              `json:"slot"`
	Ability  NamedAPIResource `json:"ability"`
}

type GameIndex struct {
	GameIndex int              `json:"game_index"`
	Version   NamedAPIResource `json:"version"`
}

type HeldItem struct {
	Item           NamedAPIResource `json:"item"`
	VersionDetails []VersionDetail  `json:"version_details"`
}

type VersionDetail struct {
	Rarity  int              `json:"rarity"`
	Version NamedAPIResource `json:"version"`
}

type Move struct {
	Move                NamedAPIResource     `json:"move"`
	VersionGroupDetails []VersionGroupDetail `json:"version_group_details"`
}

type VersionGroupDetail struct {
	LevelLearnedAt  int              `json:"level_learned_at"`
	VersionGroup    NamedAPIResource `json:"version_group"`
	MoveLearnMethod NamedAPIResource `json:"move_learn_method"`
}

type Sprites struct {
	BackDefault  string `json:"back_default"`
	BackShiny    string `json:"back_shiny"`
	FrontDefault string `json:"front_default"`
	FrontShiny   string `json:"front_shiny"`
	// Add other sprite fields as needed
}

type Stat struct {
	BaseStat int              `json:"base_stat"`
	Effort   int              `json:"effort"`
	Stat     NamedAPIResource `json:"stat"`
}

type Type struct {
	Slot int              `json:"slot"`
	Type NamedAPIResource `json:"type"`
}

// ******************** End Pokemon *******************

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
		fmt.Println("No indicated area after explore!")
		return fmt.Errorf("No indicated area after explore!")
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

func commandCatch(_ *config, pokemon string) error {

	if len(pokemon) == 0 {
		fmt.Println("No selected Pokemon to catch!")
		return fmt.Errorf("No selected Pokemon to catch!")
	}
	url := "https://pokeapi.co/api/v2/pokemon/" + pokemon

	var data Pokemon

	//byte_data, ok := cache.Get(url)
	//if ok {
	//	fmt.Printf("Using cache!") // Debug message
	//} else {
	//	fmt.Println("Cache miss! Fetching from API.") // Debug message
	//}

	res, err := http.Get(url)
	if err != nil {
		fmt.Println(fmt.Sprintf("Pokemon %s does not exist!", pokemon))
		return err
	}

	defer res.Body.Close()

	byte_data, err := io.ReadAll(res.Body)
	//err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(byte_data, &data); err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("Throwing a Pokeball at %s...", pokemon))
	base_exp := float64(data.BaseExperience)
	szansa := (szansaZlapania - base_exp/10) / 100
	if szansa >= Losuj() {
		fmt.Println(fmt.Sprintf("%s was caught!", pokemon))
		zlapane[pokemon] = data
	} else {
		fmt.Println(fmt.Sprintf("%s escaped!", pokemon))
	}
	return nil
}

func commandInspect(_ *config, pokemon string) error {
	if len(pokemon) == 0 {
		fmt.Println("You need to write what Pokemon you want to catch!")
		return fmt.Errorf("Not indicated Pokemon name")
	}
	pok, ok := zlapane[pokemon]
	if !ok {
		fmt.Println(fmt.Sprintf("You did not catch %s!", pokemon))
		return nil
	}

	fmt.Printf("Name: %s\n", pokemon)
	fmt.Printf("Height: %v\n", pok.Height)
	fmt.Printf("Weight: %v\n", pok.Weight)
	fmt.Println("Stats:")
	for _, s := range pok.Stats {
		fmt.Printf("  -%s: %v\n", s.Stat.Name, s.BaseStat)
	}
	fmt.Println("Types:")
	for _, t := range pok.Types {
		fmt.Printf("  - %s\n", t.Type.Name)
	}
	return nil
}

func commandPokedex(_ *config, _ string) error {
	fmt.Println("Your Pokedex:")
	for z, _ := range zlapane {
		fmt.Printf(" - %s\n", z)
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

// sprawdza czy udalo sie spelnic szanse
func Losuj() float64 {
	seed := time.Now().UnixNano()         // Creating a unique seed
	rng := rand.New(rand.NewSource(seed)) // Creating random numbers
	return rng.Float64()
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
