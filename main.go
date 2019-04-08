package main

import (
	"github.com/lcl101/git-sync/core"
)

func main() {
	tmp, _ := core.GetExecPath()
	conf := tmp + "sync.conf"
	app := core.App{
		ConfigPath: conf,
	}
	app.LoadConfig()
	core.Warning("app= %v", app)

	app.Sync()

}
