package bannerPlugin

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"gopkg.in/yaml.v2"
)

// Config holds server configuration settings
type Config struct {
	HammyVersion string `yaml:"hammyVersion"`
}

// LoadConfig loads the server configuration from a config.yaml file
func LoadConfig() Config {
	file, err := os.ReadFile("./config.yaml")
	if err != nil {
		log.Fatalf("Error reading config.yaml file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		log.Fatalf("Error parsing config.yaml file: %v", err)
	}

	return config
}

// PrintBanner prints the ASCII art banner with color
func PrintBanner() {
	config := LoadConfig()
	fmt.Println("\033[38;2;255;0;0m⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀ ⣀⠀⠀⠀⠀⠀⠀⠀\033[0m")
	fmt.Println("\033[38;2;255;25;0m⡠⠀⢄⡀⠀⣀⠀⠀⢀⣀⡴⠉⠀⠃⠀⠀⠀⠀⠀⠀\033[0m")
	fmt.Println("\033[38;2;255;50;0m⢇⠀⠀⣿⣾⣯⣍⣽⣿⣿⣿⡤⢀⠇⠀⠀⠀⠀⠀⠀\033[0m")
	fmt.Println("\033[38;2;255;75;0m⠀⠑⢼⣿⣿⣿⣿⣿⣿⣿⣿⣷⣷⣤⡀⠀⠀⠀⠀⠀\033[0m")
	fmt.Println("\033[38;2;255;100;0m⠀⢸⣿⣿⣟⣻⣿⣿⣿⣭⣿⣿⣿⣿⡟⠢⡀⠀⠀⠀\033[0m")
	fmt.Println("\033[38;2;255;125;0m⠀⢸⡏⢻⣿⢿⣿⣿⣿⡿⣿⡟⣿⠟⠀⠀⣿⣦⠀⠀\033[0m")
	fmt.Println("\033[38;2;255;150;0m⠀⢸⠛⠮⠝⢋⠙⣻⣊⢁⠈⠚⢃⣀⣴⣾⣿⣿⣷⡀\033[0m")
	fmt.Println("\033[38;2;255;175;0m⠀⣾⣶⣤⡠⠀⠉⠀⠈⢀⢀⣾⣿⣿⣿⣿⣿⣿⠿⣧\033[0m")
	fmt.Println("\033[38;2;255;200;0m⠀⠿⡟⠛⠻⠷⣶⠀⣶⠟⠋⠛⣿⠗⠈⠈⠉⢠⣪⣿\033[0m")
	fmt.Println("\033[38;2;255;225;0m⠀⠸⡈⠙⣄⡀⢸⢸⡿⣄⡦⠋⠁⠀⠀⠀⡠⣺⣿⣿\033[0m")
	fmt.Println("\033[38;2;255;250;0m⠀⠀⠙⢢⡤⠙⠛⡏⣅⣠⠶⠖⠒⠒⠈⠁⠐⢾⣿⡏\033[0m")
	fmt.Println("\033[38;2;255;255;0m⠀⠀⠀⠈⡄⠀⠀⠸⣿⡷⠀⠀⠀⠀⠀⢀⢠⣿⣿⠃\033[0m")
	fmt.Println("\033[38;2;255;230;0m⠀⠀⠀⠀⠘⠦⣀⠀⣿⣷⣦⠄⠀⠀⠀⢝⣿⣿⡟⠀\033[0m")
	fmt.Println("\033[38;2;255;205;0m⠀⠀⠀⣤⣤⣄⣊⡉⠟⠿⢿⡷⠗⠚⣲⠽⠿⠟⠁⠀\033[0m")
	fmt.Println("\033[38;2;255;180;0m⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠉⠉⠉⠀⠀⠀⠀⠀\033[0m")
	fmt.Printf("\033[1m\033[38;2;255;155;0mProject Hammy %s\033[0m\n", config.HammyVersion)
	fmt.Printf("\033[1m\033[38;2;255;130;0mLightning fast Go (%s) Webserver\033[0m\n", runtime.Version())
	fmt.Println("\033[1m\033[38;2;255;105;0mhttps://gohammy.org/\033[0m")
	fmt.Println("\033[0m") // Reset text color
}
