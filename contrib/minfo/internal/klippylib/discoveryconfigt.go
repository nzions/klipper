package klippylib

// the config template for printer discovery
var DiscoveryConfigTemplate = `
[mcu]
serial: {{.Port}}

[printer]
kinematics: none
max_accel: 1
max_velocity: 1
`
