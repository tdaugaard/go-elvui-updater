package main

import (
	"flag"
	"log"
	"os"
)

var (
	forceUpdate         bool
	installAddOns       bool
	worldOfWarcraftPath string
)

func init() {
	flag.BoolVar(&forceUpdate, "f", false, "force update")
	flag.BoolVar(&installAddOns, "i", false, "install add-ons if not already installed")
	flag.StringVar(&worldOfWarcraftPath, "d", "", "World of Warcraft directory")
	flag.Parse()

	if worldOfWarcraftPath == "" {
		flag.Usage()
		os.Exit(-2)
	}
}

func main() {
	updater, err := NewUpdater(worldOfWarcraftPath)
	if err != nil {
		log.Fatal(err)
	}

	updater.SetForceUpdate(forceUpdate)
	updater.SetInstallAddOns(installAddOns)
	updater.Add(NewElvUIUpdater())
	updater.Add(NewElvUAddOnSkinsIUpdater())
	updater.Check()
}
