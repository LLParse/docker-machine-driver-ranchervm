package ranchervm

import (
	"fmt"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/prometheus/common/log"
)

const (
	defaultSSHUser = "docker"
)

type Driver struct {
	*drivers.BaseDriver
	MemoryMiB int
	CPU       int
}

func NewDriver(hostName, storePath string) drivers.Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			SSHUser:     defaultSSHUser,
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

// Create a host using the driver's config
func (d *Driver) Create() error {
	// TODO
	return nil
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "ranchervm"
}

// GetCreateFlags returns the mcnflag.Flag slice representing the flags
// that can be set, their descriptions and defaults.
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	// TODO
	return []mcnflag.Flag{
		mcnflag.IntFlag{
			Name:  "ranchervm-memory-mib",
			Usage: "Memory in MiB",
			Value: 1024,
		},
		mcnflag.IntFlag{
			Name:  "ranchervm-cpu-count",
			Usage: "Number of CPUs",
			Value: 1,
		},
	}
}

// GetIP returns an IP or hostname that this host is available at
// e.g. 1.2.3.4 or docker-host-d60b70a14d3a.cloudapp.net
func (d *Driver) GetIP() (string, error) {
	// TODO
	return "", nil
}

// GetMachineName returns the name of the machine
func (d *Driver) GetMachineName() string {
	return d.MachineName
}

// GetSSHHostname returns hostname for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// GetSSHKeyPath returns key path for use with ssh
func (d *Driver) GetSSHKeyPath() string {
	return d.ResolveStorePath("id_rsa")
}

// GetSSHPort returns port for use with ssh
func (d *Driver) GetSSHPort() (int, error) {
	if d.SSHPort == 0 {
		d.SSHPort = 22
	}

	return d.SSHPort, nil
}

// GetSSHUsername returns username for use with ssh
func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "rancher"
	}

	return d.SSHUser
}

// GetURL returns a Docker compatible host URL for connecting to this host
// e.g. tcp://1.2.3.4:2376
func (d *Driver) GetURL() (string, error) {
	log.Debugf("GetURL called")
	ip, err := d.GetIP()
	if err != nil {
		log.Warnf("Failed to get IP: %s", err)
		return "", err
	}
	if ip == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s:2375", ip), nil
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	// TODO
	return state.None, nil
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	// TODO
	return nil
}

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *Driver) PreCreateCheck() error {
	// TODO
	return nil
}

// Remove a host
func (d *Driver) Remove() error {
	// TODO
	return nil
}

// Restart a host. This may just call Stop(); Start() if the provider does not
// have any special restart behaviour.
func (d *Driver) Restart() error {
	// TODO
	return nil
}

// SetConfigFromFlags configures the driver with the object that was returned
// by RegisterCreateFlags
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	log.Debugf("SetConfigFromFlags called")
	d.MemoryMiB = flags.Int("ranchervm-memory-mib")
	d.CPU = flags.Int("ranchervm-cpu-count")
	d.SSHUser = "rancher"
	d.SSHPort = 22
	return nil
}

// Start a host
func (d *Driver) Start() error {
	// TODO
	return nil
}

// Stop a host gracefully
func (d *Driver) Stop() error {
	// TODO
	return nil
}
