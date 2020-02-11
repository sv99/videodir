package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/kataras/iris/v12"
)

func InitApi(app *iris.Application, v1 iris.Party, conf *Config) {
	v1.Get("/version", func(ctx iris.Context) {
		ctx.JSON(iris.Map{"version": VERSION})
	})

	v1.Get("/volumes", func(ctx iris.Context) {
		ctx.JSON(conf.VideoDirs)
	})

	v1.Post("/list", func(ctx iris.Context) {
		var vf Path
		err := ctx.ReadJSON(&vf)
		if err != nil {
			sendJsonError(app, ctx, iris.StatusBadRequest,
				"file get ReadJSON error: " + err.Error())
			return
		}

		list := make([]string, 0, 10)

		for _, volume := range conf.VideoDirs {
			vd, err := filepath.Abs(filepath.Join(volume, filepath.Join(vf.Path...)))
			if err != nil {
				sendJsonError(app, ctx, iris.StatusInternalServerError,
					"Get full path error: " + err.Error())
				return
			}

			// this path not exists in current volume
			if _, err := os.Stat(vd); os.IsNotExist(err) {
				continue
			}

			files, err := ioutil.ReadDir(vd)
			if err != nil {
				sendJsonError(app, ctx, iris.StatusInternalServerError,
					"Read dir error: " + err.Error())
				return
			}

			for _, f := range files {
				list = append(list, f.Name())
			}
		}

		ctx.JSON(list)
	})

	v1.Post("/file", func(ctx iris.Context) {
		vf, err := readJsonPath(app, ctx)
		if err != nil {
			return
		}

		var fp = ""
		for _, volume := range conf.VideoDirs {

			fp, err = filepath.Abs(filepath.Join(volume, filepath.Join(vf.Path...)))
			if err != nil {
				sendJsonError(app, ctx, iris.StatusBadRequest,
					"Video file get full path error: " + err.Error())
				return
			}

			if _, err := os.Stat(fp); os.IsNotExist(err) {
				continue
			}
			break
		}
		ctx.SendFile(fp, filepath.Base(fp))
		app.Logger().Info("Send file: ", fp)
	})

	v1.Post("/filesize", func(ctx iris.Context) {
		vf, err := readJsonPath(app, ctx)
		if err != nil {
			return
		}

		var fp = ""
		var size int64 = 0
		for _, volume := range conf.VideoDirs {

			fp, err = filepath.Abs(filepath.Join(volume, filepath.Join(vf.Path...)))
			if err != nil {
				sendJsonError(app, ctx, iris.StatusBadRequest,
					"Video file get full path error: " + err.Error())
				return
			}

			stat, err := os.Stat(fp)
			if os.IsNotExist(err) {
				continue
			}
			size = stat.Size()
			break
		}
		ctx.JSON(iris.Map{"size": size})
		app.Logger().Info("Get file size: ", fp)
	})

	v1.Post("/remove", func(ctx iris.Context) {
		vf, err := readJsonPath(app, ctx)
		if err != nil {
			return
		}

		var fp = ""
		var res = "Ok"
		for _, volume := range conf.VideoDirs {

			fp, err = filepath.Abs(filepath.Join(volume, filepath.Join(vf.Path...)))
			if err != nil {
				sendJsonError(app, ctx, iris.StatusBadRequest,
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
			break
		}
		ctx.JSON(iris.Map{"result": res})
		app.Logger().Infof("Remove path: %s result: %s", fp, res)
	})
}
