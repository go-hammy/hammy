package main

import (
	"hammy/bannerPlugin"
	"hammy/serverPlugin"
)

var hammyVersion = "v1.01"

func main() {
	bannerPlugin.PrintBanner()
	serverPlugin.StartServer()
}