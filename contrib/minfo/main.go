package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"text/template"

	"kc/klippyclient"
)

// /dev/serial/by-id/usb-1a86_USB_Serial-if00-port0
var (
	DevDir   = "/dev/serial/by-id/"
	TempFile = "/tmp/printer.cfg"
	UDSFile  = "/tmp/klippy_uds"
)

func GetSerialPorts() (files []string, err error) {
	dl, err := ioutil.ReadDir(DevDir)
	if err != nil {
		return files, err
	}

	for _, f := range dl {
		files = append(files, f.Name())
	}
	return
}

type Printer struct {
	Name    string
	Model   string
	Port    string
	Version string
	Speed   int
}

var configT = `
[mcu]
serial: {{.Port}}

[printer]
kinematics: none
max_accel: 1
max_velocity: 1
`

func getPrinter(fName string) (p Printer, err error) {
	fmt.Println(fName)

	p.Port = filepath.Join(DevDir, fName)

	fh, err := os.Create(TempFile)
	if err != nil {
		return p, err
	}
	defer os.Remove(TempFile)

	// write config from template
	t := template.Must(template.New("config").Parse(configT))
	err = t.Execute(fh, p)
	fh.Close()
	if err != nil {
		return p, err
	}

	// start klippy api
	cmd := exec.Command("/home/pi/klippy-env/bin/python",
		"/home/pi/klipper/klippy/klippy.py",
		"/tmp/printer.cfg",
		"-a",
		UDSFile,
	)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(out))
		wg.Done()
	}()
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	// send
	kc, err := klippyclient.New(UDSFile).Dial()
	if err != nil {
		return p, err
	}

	fmt.Println(kc.GetInfo())
	fmt.Println(kc.GetMCUInfo())
	cmd.Process.Kill()

	wg.Wait()

	return p, nil
}

func main() {

	files, err := GetSerialPorts()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, f := range files {
		printer, err := getPrinter(f)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println(printer.Name)
		fmt.Println(printer.Model)
	}
}
