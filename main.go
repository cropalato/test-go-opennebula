//
// main.go
// Copyright (C) 2023 rmelo <rmelo@r-melo-lnx>
//
// Distributed under terms of the MIT license.
//

package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/joho/godotenv"
)

func getOneConfig() goca.OneConfig {
  var configPath []string
  configPath = append(configPath, "/etc")
  home, err := os.UserHomeDir() 
  if err == nil {
    configPath = append(configPath, home)
  } else {
    fmt.Printf("Error getting working dir. %s\n", err)
  }
  pwd, err := os.Getwd()
  if err == nil {
    configPath = append(configPath, pwd)
  } else {
    fmt.Printf("Error getting working dir. %s\n", err)
  }
  for _, file := range configPath {
    cfgFile := fmt.Sprintf("%s/oneEnv.cfg", file)
    err := godotenv.Load(cfgFile)
    if err != nil {
      fmt.Printf("Error reading %s.\n", cfgFile)
      continue
    }
  }
  oneUser := os.Getenv("ONE_USER")
  onePass := os.Getenv("ONE_PASS")
  oneURL := os.Getenv("ONE_URL")
  if (oneUser == "") || (onePass == "") || (oneURL == "") {
    fmt.Println("Unable to find OpenNebula credentails.")
    os.Exit(1)
  }
  return goca.NewConfig(oneUser, onePass, oneURL)
}

func main() {
  fmt.Println("Staring ...")
  client := goca.NewDefaultClient(
    getOneConfig(),
  )
  controller := goca.NewController(client)
  vmController := controller.VMs()
  vmPool, err := vmController.InfoExtended(-2, -1, -1, -1) // ref: https://docs.opennebula.io/6.4/integration_and_development/system_interfaces/api.html look for one.vmpool.infoextended
  if err != nil {
    fmt.Printf("Failed calling for VM info. %s\n", err)
    os.Exit(1)
  }
  w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
  for _, vm := range vmPool.VMs {
    s1, s2, err := vm.StateString()
  if err != nil {
    fmt.Printf("Failed getting VM state. %s\n", err)
    os.Exit(1)
  }
    fmt.Fprintln(w, fmt.Sprintf("%d\t%s\t%s(%s)\t", vm.ID, vm.Name, s1, s2))
  }
  w.Flush()
}
