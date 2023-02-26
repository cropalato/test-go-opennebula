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
  "regexp"
  "text/tabwriter"

  "github.com/OpenNebula/one/src/oca/go/src/goca"
  //"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
  //"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
  //"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm/keys"
  "github.com/joho/godotenv"
)

func loadEnvFile() {
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
}

func getOneConfig() goca.OneConfig {
  oneUser := os.Getenv("ONE_USER")
  onePass := os.Getenv("ONE_PASS")
  oneURL := os.Getenv("ONE_URL")
  if (oneUser == "") || (onePass == "") || (oneURL == "") {
    fmt.Println("Unable to find OpenNebula credentials.")
    os.Exit(1)
  }
  return goca.NewConfig(oneUser, onePass, oneURL)
}

func main() {
  fmt.Println("Staring ...")
  loadEnvFile()
  client := goca.NewDefaultClient(
    getOneConfig(),
  )
  controller := goca.NewController(client)



  // Listing all VMs
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

  fmt.Printf("\n\n\n\n")

  //Listing all images
  imgPattern := os.Getenv("ONE_VM_IMGPATTERN")
  re := regexp.MustCompile(fmt.Sprintf("%s",imgPattern))
  imgController := controller.Images()
  imgPool, err := imgController.Info(-2, -1, -1)
  if err != nil {
    fmt.Printf("Failed calling for images info. %s\n", err)
    os.Exit(1)
  }
  newerImg := -1
  w = tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
  for _, img := range imgPool.Images {
    imgID := img.ID
    imgName := img.Name
    imgType := img.Type // 0 = OS, 1 = CDROM, 2 = DATABLOCK, 5 = file
    if imgType == "0" && re.Match([]byte(imgName)) {
      if imgID > newerImg { newerImg = imgID }
      fmt.Fprintln(w, fmt.Sprintf("%d\t%s\t%s\t", imgID, imgName, imgType))
    }
  }
  w.Flush()

  fmt.Printf("We will use image %d.\n", newerImg)

  //Getting Network info
  netName := os.Getenv("ONE_VM_NET")
  netController := controller.VirtualNetworks()
  netID, err := netController.ByName(netName, -2, -1, -1)
  if err != nil {
    fmt.Printf("Error getting net info for %s. %s\n", netName, err)
    os.Exit(1)
  }
  fmt.Printf("NetId=%d.\n", netID)

/*
  //Creating a new VM
  // Build a string template. (No XML-RPC call done)
  // To make a VM from an existing OpenNebula template,
  // use template "Instantiate" method instead
  tpl := vm.NewTemplate()
  tpl.Add(keys.Name, "this-is-a-vm")
  tpl.CPU(1).Memory(64).VCPU(2)

  // The image ID should exist to make this example work
  disk := tpl.AddDisk()
  disk.Add(shared.ImageID, newerImg)
  disk.Add(shared.DevPrefix, "vd")

  // The network ID should exist to make this example work
  nic := tpl.AddNIC()
  nic.Add(shared.NetworkID, netID)
  nic.Add(shared.Model, "virtio")

  // Create VM from template
  vmID, err := controller.VMs().Create(tpl.String(), false)
  if err != nil {
    fmt.Printf("Error creating new VM. %s\n", err)
    os.Exit(1)
  }

  vmCtrl := controller.VM(vmID)

  // Fetch informations of the created VM
  vm, err := vmCtrl.Info(false)
  if err != nil {
    fmt.Printf("Error getting VM Info. %s\n", err)
    os.Exit(1)
  }

  fmt.Printf("%+v\n", vm)

  // Poweroff the VM
  err = vmCtrl.Poweroff()
  if err != nil {
    fmt.Printf("Error stoping VM. %s\n", err)
    os.Exit(1)
  }
*/
}
