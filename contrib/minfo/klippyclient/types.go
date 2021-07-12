package klippyclient

type Command struct {
	ID     int         `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

type CommandObjectList struct {
	Objects map[string][]string `json:"objects"`
}

type Response struct {
	ID     int         `json:"id"`
	Result interface{} `json:"result,omitempty"`
}

type InfoResponse struct {
	StateMessage    string `json:"state_message"`
	KlipperPath     string `json:"klipper_path"`
	ConfigFile      string `json:"config_file"`
	SoftwareVersion string `json:"software_version"`
	Hostname        string `json:"hostname"`
	CPUInfo         string `json:"cpu_info"`
	State           string `json:"state"`
	PythonPath      string `json:"python_path"`
	LogFile         string `json:"log_file"`
}

type MCUResponse struct {
	Status struct {
		Mcu struct {
			McuBuildVersions string `json:"mcu_build_versions"`
			McuVersion       string `json:"mcu_version"`
			McuConstants     struct {
				MachineName   string `json:"machine_name"`
				MachineModel  string `json:"machine_model"`
				ReceiveWindow int    `json:"RECEIVE_WINDOW"`
				StepDelay     int    `json:"STEP_DELAY"`
				SerialBaud    int    `json:"SERIAL_BAUD"`
				AdcMax        int    `json:"ADC_MAX"`
				PwmMax        int    `json:"PWM_MAX"`
				Mcu           string `json:"MCU"`
				ClockFreq     int    `json:"CLOCK_FREQ"`
			} `json:"mcu_constants"`
		} `json:"mcu"`
	} `json:"status"`
	Eventtime float64 `json:"eventtime"`
}
