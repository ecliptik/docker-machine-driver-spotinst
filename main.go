package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/docker-machine-driver-spotinst/spotinst"
)

func main() {
	plugin.RegisterDriver(spotinst.NewDriver("", ""))
}