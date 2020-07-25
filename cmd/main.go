package main

import (
	"log"
	"os"
	"videodir"
)

func main() {
	conf := videodir.DefaultConfig()
	err := conf.TOML("./videodir.conf")
	if err != nil {
		log.Panicf("Config load problems: %s", err.Error())
	}
	app := videodir.NewApp(&conf)
	// dump config
	app.Logger.Debugf("Cert: %s", app.Config.Cert)
	app.Logger.Debugf("Key: %s", app.Config.Key)
	app.Logger.Debugf("Videodirs: %s", app.Config.VideoDirs)

	cliApp := videodir.InitCli(app)
	cliApp.WithAction(func(args []string, options map[string]string) int {

		//app.Run(iris.TLS(conf.ServerAddr, conf.Cert, conf.Key))
		//return 0

		app.Logger.Infof("Server running on https://localhost:%s", app.Config.ServerAddr)
		app.Serve()
		return 0
	})

	// Start server
	os.Exit(cliApp.Run(os.Args, os.Stdout))
}
