package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/rancher/docker-machine-driver-ranchervm"
)

func main() {
	plugin.RegisterDriver(ranchervm.NewDriver("default", "path"))
}
