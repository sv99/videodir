// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package main

import (
	"videodir/videodir"
	"log"
	"os"
)

// This is the name you will use for the NET START command
const svcName = "gosvc"

// This is the name that will appear in the Services control panel
const svcNameLong = "GO Service"

func svcLauncher() error {

	conf := videodir.DefaultConfig()
	err := conf.TOML("videodir.conf")
	if err != nil {
		log.Panicf("Config load problems: %s", err.Error())
	}
	app := videodir.NewApp(&conf)

	cliApp := videodir.InitCli(app)
	cliApp.WithAction(func(args []string, options map[string]string) int {
		app.Logger.Infof("Server running on https://localhost:%s", app.Config.ServerAddr)
		app.Serve()
		return 0
	})

	// Start server
	go cliApp.Run(os.Args, os.Stdout)
	return nil
}
