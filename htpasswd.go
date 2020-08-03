package videodir

import (
	"fmt"
	"path/filepath"

	"github.com/foomo/htpasswd"
)

func GetHtPasswdPath(workDir string) string {
	return filepath.Join(workDir, HTPASSWD)
}

func ListUsers(workDir string) error {
	pass, err := htpasswd.ParseHtpasswdFile(GetHtPasswdPath(workDir))
	if err != nil {
		return err
	}
	for name := range pass {
		fmt.Println(name)
	}
	return nil
}

func AddUser(workDir, name, pass string) error {
	return htpasswd.SetPassword(GetHtPasswdPath(workDir), name, pass, htpasswd.HashBCrypt)
}

func RemoveUser(workDir, name string) error {
	return htpasswd.RemoveUser(GetHtPasswdPath(workDir), name)
}
