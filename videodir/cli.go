package videodir

import (
	"fmt"
	"github.com/foomo/htpasswd"
	"github.com/teris-io/cli"
)

func InitCli(app *AppServer) cli.App {
	list := cli.NewCommand("list", "list users from htpasswd").
		WithAction(func(args []string, options map[string]string) int {
		for name, _ := range app.passwords {
			fmt.Println(name)
		}
		return 0
	})

	add := cli.NewCommand("add", "add or update user in the htpasswd").
		WithArg(cli.NewArg("name", "user name")).
		WithArg(cli.NewArg("password", "password")).
		WithAction(func(args []string, options map[string]string) int {
		err := htpasswd.SetPassword(HTPASSWD, args[0], args[1], htpasswd.HashBCrypt)
		if err != nil {
			app.Logger.Fatal("update htpasswd error: ", err.Error())
		}
		fmt.Println("User added: ", args[0])
		return 0
	})

	remove := cli.NewCommand("remove", "remove user from htpasswd").
		WithArg(cli.NewArg("name", "user name")).
		WithAction(func(args []string, options map[string]string) int {
		err := htpasswd.RemoveUser(HTPASSWD, args[0])
		if err != nil {
			app.Logger.Fatal("remove user from htpasswd error: ", err.Error())
		}
		fmt.Println("User removed: ", args[0])
		return 0
	})

	return cli.New("videodir tool").
		WithCommand(list).
		WithCommand(add).
		WithCommand(remove)
}
