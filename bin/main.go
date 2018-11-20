package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	ranchervm "github.com/llparse/docker-machine-driver-ranchervm"
)

func main() {
	plugin.RegisterDriver(ranchervm.NewDriver("", ""))
}
