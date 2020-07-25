package videodir

import (
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type Path struct {
	Path []string `json:"path"`
}

func (srv *AppServer)  readPath(c *fiber.Ctx) (*Path, error) {
	var vf Path
	err := json.Unmarshal([]byte(c.Body()), &vf)
	if err != nil {
		log.Printf("File get ReadJSON error: %s %s", c.Body(), err.Error())
		c.Status(fiber.StatusBadRequest)
		return &vf, err
	}

	// empty path not available
	if len(vf.Path) == 0 {
		log.Printf("File path not specified")
		c.Status(fiber.StatusBadRequest)
		return &vf, errors.New("file path not specified")
	}
	return &vf, nil
}

func (srv *AppServer) ListPath(c *fiber.Ctx) {
	var vf Path
	err := json.Unmarshal([]byte(c.Body()), &vf)
	if err != nil {
		log.Printf("File get ReadJSON error: %s %s", c.Body(), err.Error())
		c.Status(fiber.StatusBadRequest)
		return
	}

	list := make([]string, 0, 10)

	for _, volume := range srv.Config.VideoDirs {
		vd, err := filepath.Abs(filepath.Join(volume, filepath.Join(vf.Path...)))
		if err != nil {
			srv.Error(c, fiber.StatusInternalServerError,
				"Get full path error: " + err.Error())
			return
		}

		// this path not exists in current volume
		if _, err := os.Stat(vd); os.IsNotExist(err) {
			continue
		}

		files, err := ioutil.ReadDir(vd)
		if err != nil {
			srv.Error(c, fiber.StatusInternalServerError,
				"Read dir error: " + err.Error())
			return
		}

		for _, f := range files {
			list = append(list, f.Name())
		}
	}
	_ = c.JSON(list)
}

func (srv *AppServer) PostFile(c *fiber.Ctx) {
	vf, err := srv.readPath(c)
	if err != nil {
		return
	}

	var fp = ""
	for _, volume := range srv.Config.VideoDirs {

		fp, err = filepath.Abs(filepath.Join(volume, filepath.Join(vf.Path...)))
		if err != nil {
			srv.Error(c, fiber.StatusInternalServerError,
				"Video file get full path error: " + err.Error())
			return
		}

		if _, err := os.Stat(fp); os.IsNotExist(err) {
			continue
		} else {
			_ = c.SendFile(fp)
			srv.Logger.Warnf("Send file: %s", fp)
			return
		}
	}
	srv.Error(c, fiber.StatusBadRequest,
		"Video file not found")

}

func (srv *AppServer) PostFileSize(c *fiber.Ctx) {
	vf, err := srv.readPath(c)
	if err != nil {
		return
	}

	var fp = ""
	var size int64 = 0
	for _, volume := range srv.Config.VideoDirs {

		fp, err = filepath.Abs(filepath.Join(volume, filepath.Join(vf.Path...)))
		if err != nil {
			srv.Error(c, fiber.StatusBadRequest,
				"Video file get full path error: " + err.Error())
			return
		}

		stat, err := os.Stat(fp)
		if os.IsNotExist(err) {
			continue
		}
		size = stat.Size()
		_ = c.JSON(fiber.Map{"size": size})
		srv.Logger.Warnf("Get file size: %s", fp)
		return
	}
	srv.Error(c, fiber.StatusBadRequest,
		"Video file not found")
}

// remove file or directory
// find path for remove on all volumes
func (srv *AppServer) RemoveFile(c *fiber.Ctx) {
	vf, err := srv.readPath(c)
	if err != nil {
		return
	}

	var fp = ""
	var res = "Ok"
	for _, volume := range srv.Config.VideoDirs {

		fp, err = filepath.Abs(filepath.Join(volume, filepath.Join(vf.Path...)))
		if err != nil {
			srv.Error(c, fiber.StatusBadRequest,
				"Video file get full path error: " + err.Error())
			return
		}

		if _, err := os.Stat(fp); os.IsNotExist(err) {
			continue
		}

		err = os.RemoveAll(fp)
		if err != nil {
			res = err.Error()
		}
	}
	_ = c.JSON(fiber.Map{"result": res})
	srv.Logger.Warnf("Remove path: %s result: %s", fp, res)
}

