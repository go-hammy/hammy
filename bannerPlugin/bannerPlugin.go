package bannerPlugin

import (
	"fmt"
	"runtime"
)

// PrintBanner prints the ASCII art banner with color
func PrintBanner() {
	fmt.Println("\033[38;2;255;0;0m⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⠀⠀⠀⠀⠀⠀⠀\033[0m")
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
	fmt.Println("\033[1m\033[38;2;255;155;0mProject Hammy\033[0m")
	fmt.Println("\033[1m\033[38;2;255;130;0mLightning fast Go (" + fmt.Sprintf("%s", runtime.Version()) + ") Webserver\033[0m")
	fmt.Println("\033[1m\033[38;2;255;105;0mhttps://ricardomolendijk.com/\033[0m")
	fmt.Println("\033[0m") // Reset text color
}
