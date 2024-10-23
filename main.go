package main

import (
	"hammy/bannerPlugin"
	"hammy/serverPlugin"
)

func main() {
	bannerPlugin.PrintBanner()
	serverPlugin.StartServer()
}