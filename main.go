package main


import (
  "./gopev"
  "io/ioutil"
  "github.com/fatih/color"
  "log"
  "os"
)

func main() {
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
