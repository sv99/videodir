// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"golang.org/x/sys/windows/svc"

	"videodir"
)

func main()  {
	// This is the name you will use for the NET START command
	const svcName = "videodir"
	// This is the name that will appear in the Services control panel
	const svcDesc = "VideoDir Service"

	// Working directory for windows - exe directory!!
	ex, err := os.Executable()
	if err != nil {
		fmt.Errorf("get executable %v", err)
		return
	}
	workDir := filepath.Dir(ex)

	// service initialize code
	isIntSess, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	}

	service := videodir.NewService(svcName, svcDesc, isIntSess)

	// check service mode
	if !isIntSess {
		service.Run()
		return
	}

	cliApp := &cli.App{
		Name: "videodir",
		Usage: "video registrator storage backend",
		//Action: func(c *cli.Context) error {
		//	fmt.Println("Hello friend!")
		//	return nil
		//},
		Commands: []*cli.Command{
			{
				Name:    "install",
				Usage:   "install service",
				Action:  func(c *cli.Context) error {
					err := service.Install()
					if err != nil {
						return fmt.Errorf("install error %v", err)
					}
					fmt.Println("service installed")
					return nil
				},
			},
			{
				Name:    "remove",
				Usage:   "remove service",
				Action:  func(c *cli.Context) error {
					err := service.Remove()
					if err != nil {
						return fmt.Errorf("remove error %v", err)
					}
					fmt.Println("service removed")
					return nil
				},
			},
			{
				Name:    "start",
				Usage:   "start service",
				Action:  func(c *cli.Context) error {
					err := service.Start()
					if err != nil {
						return fmt.Errorf("start error %v", err)
					}
					fmt.Println("service started")
					return nil
				},
			},
			{
				Name:    "stop",
				Usage:   "stop service",
				Action:  func(c *cli.Context) error {
					err := service.Control(svc.Stop, svc.Stopped)
					if err != nil {
						return fmt.Errorf("stop error %v", err)
					}
					fmt.Println("service stopped")
					return nil
				},
			},
			{
				Name:    "debug",
				Usage:   "debug service",
				Action:  func(c *cli.Context) error {
					return service.Run()
				},
			},
			{
				Name:        "users",
				Usage:       "manage htpasswd",
				Subcommands: []*cli.Command{
					{
						Name:  "list",
						Usage: "list users",
						Action: func(c *cli.Context) error {
							return videodir.ListUsers(workDir)
						},
					},
					{
						Name:  "add",
						Usage: "add or update user",
						Action: func(c *cli.Context) error {
							return videodir.AddUser(workDir, c.Args().First(),c.Args().Get(1) )
						},
					},
					{
						Name:  "remove",
						Usage: "remove user",
						Action: func(c *cli.Context) error {
							return videodir.RemoveUser(workDir, c.Args().First())
						},
					},
				},
			},
		},
	}

	err = cliApp.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
