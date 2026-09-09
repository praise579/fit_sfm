package main

import (
	_ "embed"

	"github.com/praise579/fit_sfm/cmd"
)

//go:embed config/config.yaml
var c string

func main() {
	cmd.Execute(c)
}
