package milib

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"minfo/local/klippylib"
	"minfo/local/kp"
	"path/filepath"
	"time"

	"github.com/nzions/toolbox/x/baselog"
)

var (
	DeviceDir = "/dev/serial/by-id/"

	RetryCount = 3
	RetryDelay = time.Second * 5
)

func ListSerialPorts() (files []string, err error) {
	dl, err := ioutil.ReadDir(DeviceDir)
	if err != nil {
		return files, err
	}

	for _, f := range dl {
		files = append(files, f.Name())
		baselog.DefaultLogger.Logf("found serial port %s", f.Name())
	}
	return
}

func DiscoverPrinters() (printers []kp.Printer, err error) {
	ports, err := ListSerialPorts()
	if err != nil {
		return printers, err
	}

	for _, prt := range ports {
		p, err := DiscoverPrinter(prt)
		if err != nil {
			baselog.DefaultLogger.Errorf("error discovering %s: %s", prt, err)
			continue
		}

		printers = append(printers, p)
	}
	return

}

func DiscoverPrinter(port string) (printer kp.Printer, err error) {
	md5 := md5.Sum([]byte(port))
	uid := fmt.Sprintf("%x", md5[12:])

	printer.PortShort = port
	printer.Port = filepath.Join(DeviceDir, port)
	printer.UID = uid

	baselog.DefaultLogger.Debugf("[%s] starting discovery on %s", printer.UID, printer.PortShort)

	// create a server
	d, err := klippylib.NewServer(klippylib.NewServerConfig(uid))
	if err != nil {
		return printer, err
	}
	defer d.Stop()

	if err := d.WriteDiscoveyConfig(printer); err != nil {
		return printer, err
	}

	if err := d.StartServer(); err != nil {
		return printer, err
	}

	baselog.DefaultLogger.Logf("[%s] waiting 2 secs so klippy can start and connect", printer.UID)
	time.Sleep(time.Millisecond * 2000)

	mcuInfo, err := GetMCUInfo(d.Config)
	if err != nil {
		d.DumpLogs = true
		return printer, err
	}

	printer.Name = mcuInfo.Status.Mcu.McuConstants.MachineName
	printer.Model = mcuInfo.Status.Mcu.McuConstants.MachineModel
	printer.Version = mcuInfo.Status.Mcu.McuVersion
	baselog.DefaultLogger.Logf("[%s] found %s the %s on %s", printer.UID, printer.Name, printer.Model, printer.PortShort)
	return
}

func getMCUInfo(config klippylib.ServerConfig) (resp klippylib.MCUResponse, err error) {
	kc, err := klippylib.NewClient(config.UDSFile).Dial()
	if err != nil {
		return resp, err
	}
	defer kc.Close()

	baselog.DefaultLogger.Debugf("[%s] connected to klippy", config.UID)

	info, err := kc.GetInfo()
	if err != nil {
		return resp, err
	}
	if info.State != "ready" {
		return resp, fmt.Errorf("state %s", info.State)
	}
	baselog.DefaultLogger.Debugf("[%s] printer online", config.UID)

	return kc.GetMCUInfo()
}

// a convienience wrapper to get mcuinfo
func GetMCUInfo(config klippylib.ServerConfig) (resp klippylib.MCUResponse, err error) {

	// we do this all in a giant loop as klippy kills connections when a command has an error
	// e.g. getinfo closes when the printer aint ready
	for i := 0; i < RetryCount; i++ {
		resp, err = getMCUInfo(config)
		if err == nil {
			break
		}
		baselog.DefaultLogger.Debugf("[%s] [%d/%d] %s", config.UID, i+1, RetryCount, err)
		time.Sleep(RetryDelay)
	}
	if err != nil {
		return resp, fmt.Errorf("[%s] Failed discovering printer: %s", config.UID, err.Error())
	}
	return
}
