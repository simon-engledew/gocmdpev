package main

import (
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/simon-engledew/gocmdpev/pev"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("gocmdpev", "A command-line GO Postgres query visualizer (see https://github.com/simon-engledew/gocmdpev).")
)

func main() {
	app.HelpFlag.Short('h')
	app.Version("1.0.0")
	app.VersionFlag.Short('v')
	app.Parse(os.Args[1:])

	err := pev.Visualize(color.Output, os.Stdin)

	if err != nil {
		log.Fatalf("%v", err)
	}
}
