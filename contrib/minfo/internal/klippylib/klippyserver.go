package klippylib

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"minfo/internal/kp"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"time"

	"github.com/nzions/toolbox/x/baselog"
)

var (
	TempDir    = "/tmp"
	PythonPath = "/home/pi/klippy-env/bin/python"
	KlippyPath = "/home/pi/klipper/klippy/klippy.py"
)

func NewServerConfig(uid string) (config ServerConfig) {
	config.UID = uid
	config.ConfigFile = fmt.Sprintf("%s/ksvr_%s_printer.cfg", TempDir, uid)
	config.UDSFile = fmt.Sprintf("%s/ksvr_%s_uds", TempDir, uid)
	config.RunFile = fmt.Sprintf("%s/ksvr_%s.run", TempDir, uid)

	return
}

type ServerConfig struct {
	ConfigFile string
	UDSFile    string
	RunFile    string
	UID        string
}

func NewServer(c ServerConfig) (*Server, error) {
	svr := &Server{
		Config: c,
	}
	return svr, svr.Check()
}

type Server struct {
	Config   ServerConfig
	DumpLogs bool

	cmd        *exec.Cmd
	wg         sync.WaitGroup
	stopCalled bool
}

func (x *Server) Check() error {
	if _, err := os.Stat(x.Config.RunFile); !os.IsNotExist(err) {
		return fmt.Errorf("another klippy is running")
	}
	return ioutil.WriteFile(x.Config.RunFile, []byte("0"), 0755)
}

func (x *Server) StartServer() error {
	baselog.DefaultLogger.Logf("[%s] starting klippy", x.Config.UID)
	x.cmd = exec.Command(PythonPath,
		KlippyPath,
		x.Config.ConfigFile,
		"-a",
		x.Config.UDSFile,
	)

	// start command in goroutine
	x.wg.Add(1)
	go func() {

		eStr := ""
		out, err := x.cmd.CombinedOutput()
		if err != nil {
			eStr = err.Error()
		}
		if x.stopCalled {
			baselog.DefaultLogger.Debugf("[%s] klippy died a natural death", x.Config.UID)
		} else {
			baselog.DefaultLogger.Debugf("[%s] klippy exited: %s", x.Config.UID, eStr)
			x.DumpLogs = true
		}

		if x.DumpLogs {
			baselog.DefaultLogger.Logf("[%s] Logs from klippy ----- ", x.Config.UID)
			fmt.Println(string(out))
			baselog.DefaultLogger.Logf("[%s] End Logs from klippy ----- ", x.Config.UID)
		}

		x.wg.Done()
	}()

	var running bool
	var pid int
	for !running {
		running, pid = x.Status()
		time.Sleep(time.Millisecond * 20)
	}
	baselog.DefaultLogger.Debugf("[%s] klippy started pid: %d", x.Config.UID, pid)

	// now catch ctrl+c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			baselog.DefaultLogger.Logf("[%s] %s: please wait for cleanup", x.Config.UID, sig)
			x.Stop()
		}
	}()

	return nil
}

func (x *Server) Stop() error {
	baselog.DefaultLogger.Debugf("[%s] stopping klippy", x.Config.UID)
	x.stopCalled = true

	if x.cmd.Process != nil {
		baselog.DefaultLogger.Debugf("[%s] killing klippy pid: %d", x.Config.UID, x.cmd.Process.Pid)
		x.cmd.Process.Kill()
		x.wg.Wait()
	}

	// remove files
	for _, f := range []string{x.Config.RunFile, x.Config.UDSFile} {
		if _, err := os.Stat(f); !os.IsNotExist(err) {
			baselog.DefaultLogger.Debugf("[%s] cleaning up %s", x.Config.UID, f)
			os.Remove(f)
		}
	}
	baselog.DefaultLogger.Logf("[%s] klippy has stopped", x.Config.UID)
	return nil
}

func (x *Server) Status() (running bool, pid int) {
	// processState is created once a process exits
	if x.cmd.ProcessState != nil {
		return false, 0
	}

	// cmd.process is created once a process has started
	if x.cmd.Process == nil {
		return false, 0
	}

	// we running, return the pid
	return true, x.cmd.Process.Pid
}

func (x *Server) WriteDiscoveyConfig(printer kp.Printer) error {
	baselog.DefaultLogger.Debugf("[%s] writing discovery config to %s", x.Config.UID, x.Config.ConfigFile)

	fh, err := os.Create(x.Config.ConfigFile)
	if err != nil {
		return err
	}
	defer fh.Close()

	// write config from template
	t := template.Must(template.New("config").Parse(DiscoveryConfigTemplate))
	return t.Execute(fh, printer)
}
