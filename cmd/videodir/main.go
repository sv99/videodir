package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"videodir"
)

func main() {
	// Working directory for nix - pwd()
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	logger, err := videodir.NewLogger(workDir, true)
	app := videodir.NewApp(filepath.Join(workDir, videodir.CONF_FILE), logger)

	cliApp := &cli.App{
		Name: "videodir",
		Usage: "video registrator storage backend",
		Action: func(c *cli.Context) error {
			app.Serve()
			return nil
		},
		Commands: []*cli.Command{
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
	}

	err = cliApp.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
