package main

import (
  "github.com/simon-engledew/gocmdpev/gopev"
  "gopkg.in/alecthomas/kingpin.v2"
  "io/ioutil"
  "github.com/fatih/color"
  "log"
  "os"
)

func main() {
  kingpin.CommandLine.HelpFlag.Short('h')
  kingpin.CommandLine.Version("1.0.0")
  kingpin.CommandLine.VersionFlag.Short('v')
  kingpin.Parse()

  buffer, err := ioutil.ReadAll(os.Stdin)

  if err != nil {
    log.Fatalf("%v", err)
  }

  // fmt.Println(string(buffer))

  err = gopev.Visualize(color.Output, buffer)

  if err != nil {
    log.Fatalf("%v", err)
  }
}
