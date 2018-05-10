package main

import (
	"github.com/docker-machine-driver-spotinst/spotinst"
	"github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
	plugin.RegisterDriver(spotinst.NewDriver("", ""))
}
