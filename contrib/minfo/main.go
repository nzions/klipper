package main

import (
	"fmt"
	"minfo/internal/milib"
	"os"
	"os/signal"
	"time"

	"github.com/nzions/toolbox/x/baselog"
)

func main() {
	baselog.DefaultLogger.SetLogLevel(baselog.LogLevelLog)

	// kinda hacky
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			fmt.Println("")
			baselog.DefaultLogger.Logf("ctrl+c: giving the servers a second to exit...")
			time.Sleep(time.Second * 1)
			os.Exit(1)
		}
	}()

	printers, err := milib.DiscoverPrinters()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("")
	fmt.Println("Discovered Printers")
	fStr := "%-15s %-20s %-35s %s\n"
	fmt.Printf(fStr, "Name", "Model", "Port", "Version")
	for _, p := range printers {
		fmt.Printf(fStr, p.Name, p.Model, p.PortShort, p.Version)
	}
	os.Exit(0)
}
