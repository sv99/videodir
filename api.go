package videodir

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

type Path struct {
	Path []string `json:"path"`
}

func (srv *AppServer) readPath(c *fiber.Ctx) (*Path, error) {
	var vf Path
	err := json.Unmarshal(c.Body(), &vf)
	if err != nil {
		srv.Logger.Error().Msgf("File get ReadJSON error: %s %v", c.Body(), err)
		c.Status(fiber.StatusBadRequest)
		return &vf, err
	}

	// empty path not available
	if len(vf.Path) == 0 {
		srv.Logger.Error().Msg("File path not specified")
		c.Status(fiber.StatusBadRequest)
		return &vf, errors.New("file path not specified")
	}
	return &vf, nil
}

func (srv *AppServer) ListPath(c *fiber.Ctx) error {
	var vf Path
	err := json.Unmarshal(c.Body(), &vf)
	if err != nil {
		srv.Logger.Error().Msgf("File get ReadJSON error: %s %v", c.Body(), err)
		c.Status(fiber.StatusBadRequest)
		return err
	}

	list := make([]string, 0, 10)

	for _, volume := range srv.Config.VideoDirs {
		vd, err := filepath.Abs(filepath.Join(volume, filepath.Join(vf.Path...)))
		if err != nil {
			_ = srv.Error(c, fiber.StatusInternalServerError,
				"Get full path error: "+err.Error())
			return err
		}
		srv.Logger.Info().Msgf("volume full path %s", vd)
		// this path not exists in current volume
		if _, err := os.Stat(vd); os.IsNotExist(err) {
			srv.Logger.Info().Msgf("volume stat err %v", err)
			continue
		}

		files, err := ioutil.ReadDir(vd)
		if err != nil {
			_ = srv.Error(c, fiber.StatusInternalServerError,
				"Read dir error: "+err.Error())
			return err
		}

		for _, f := range files {
			list = append(list, f.Name())
		}
	}
	return c.JSON(list)
}

func (srv *AppServer) PostFile(c *fiber.Ctx) error {
	vf, err := srv.readPath(c)
	if err != nil {
		return err
	}

	var fp = ""
	for _, volume := range srv.Config.VideoDirs {

		fp, err = filepath.Abs(filepath.Join(volume, filepath.Join(vf.Path...)))
		if err != nil {
			_ = srv.Error(c, fiber.StatusInternalServerError,
				"Video file get full path error: "+err.Error())
			return err
		}

		if _, err := os.Stat(fp); os.IsNotExist(err) {
			continue
		} else {
			err = c.SendFile(fp)
			srv.Logger.Warn().Msgf("Send file: %s", fp)
			return err
		}
	}
	return srv.Error(c, fiber.StatusBadRequest,
		"Video file not found")
}

func (srv *AppServer) PostFileSize(c *fiber.Ctx) error {
	vf, err := srv.readPath(c)
	if err != nil {
		return err
	}

	var fp = ""
	var size int64 = 0
	for _, volume := range srv.Config.VideoDirs {

		fp, err = filepath.Abs(filepath.Join(volume, filepath.Join(vf.Path...)))
		if err != nil {
			_ = srv.Error(c, fiber.StatusBadRequest,
				"Video file get full path error: "+err.Error())
			return err
		}

		stat, err := os.Stat(fp)
		if os.IsNotExist(err) {
			continue
		}
		size = stat.Size()
		_ = c.JSON(fiber.Map{"size": size})
		srv.Logger.Warn().Msgf("Get file size: %s", fp)
		return err
	}
	return srv.Error(c, fiber.StatusBadRequest,
		"Video file not found")
}

// RemoveFile remove file or directory
// find path for remove on all volumes
func (srv *AppServer) RemoveFile(c *fiber.Ctx) error {
	vf, err := srv.readPath(c)
	if err != nil {
		return err
	}

	var fp = ""
	var res = "Ok"
	for _, volume := range srv.Config.VideoDirs {

		fp, err = filepath.Abs(filepath.Join(volume, filepath.Join(vf.Path...)))
		if err != nil {
			_ = srv.Error(c, fiber.StatusBadRequest,
				"Video file get full path error: "+err.Error())
			return err
		}

		if _, err := os.Stat(fp); os.IsNotExist(err) {
			continue
		}

		err = os.RemoveAll(fp)
		if err != nil {
			res = err.Error()
		}
	}
	srv.Logger.Warn().Msgf("Remove path: %s result: %s", fp, res)
	return c.JSON(fiber.Map{"result": res})
}
