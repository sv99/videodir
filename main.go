package main

import (
	"os"
	"strconv"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	//"github.com/dgrijalva/jwt-go"
	//jwtmiddleware "github.com/iris-contrib/middleware/jwt"
)

const VERSION = "0.1"

type Config struct {
	ServerAddr string
	Cert       string
	Key        string
	VideoDirs  []string
	VideoDirsMap map[string]string
}

type VideoFile struct {
	VolumeId 	string `json:"volumeid"`
	Path 	string `json:"path"`
}

func main() {

	//run server
	app := iris.New()
	app.Logger().SetLevel("debug")

	app.Use(recover.New())
	app.Use(logger.New())

	//get path name for the executable
	//ex, err := os.Executable()
	//if err != nil {
	//	app.Logger().Warn(err)
	//	panic(err)
	//}
	//exPath := path.Dir(ex)

	//read configuration
	conf := Config{
		ServerAddr: ":8443",
		Cert:       "server.crt",
		Key:        "server.key",
	}
	if _, err := toml.DecodeFile("./videodir.conf", &conf); err != nil {
		app.Logger().Warn("Config problems: " + err.Error())
	}

	// init conf.VideoDirsMap
	conf.VideoDirsMap = make(map[string]string)

	app.Logger().Info("key: ", conf.Key)
	app.Logger().Info("cert: ", conf.Cert)
	for i, s := range conf.VideoDirs {
		conf.VideoDirsMap[strconv.Itoa(i)] = s
		app.Logger().Info("videoDir[", strconv.Itoa(i), "] = ", s)
	}

	app.Favicon("./favicon.ico")

	app.RegisterView(iris.HTML("./", ".html"))
	app.Get("/", func(ctx iris.Context) {
		ctx.View("index.html")
	})

	//app.StaticWeb("/video1", *videoDir1)

	v1 := app.Party("/api/v1")
	{
		v1.Get("/version", func(ctx iris.Context) {
			ver := iris.Map{"version": VERSION}
			ctx.JSON(ver)
			ver = nil
		})

		v1.Get("/list", func(ctx iris.Context) {
			ctx.JSON(conf.VideoDirsMap)
		})

		v1.Get("/list/{volume}", func(ctx iris.Context) {
			volumeId := ctx.Params().Get("volume")
			videoDir := conf.VideoDirsMap[volumeId]
			if videoDir == "" {
				ctx.StatusCode(iris.StatusBadRequest)
				ctx.JSON(iris.Map{"status": iris.StatusBadRequest, "error": "volume" + volumeId})
				return
			}

			vd, err := filepath.Abs(videoDir)
			vdLen := len(vd)
			if err != nil {
				app.Logger().Error("Get full path error", videoDir)
				ctx.StatusCode(iris.StatusBadRequest)
				ctx.JSON(iris.Map{"status": iris.StatusBadRequest, "error": vd})
				return
			}

			if _, err := os.Stat(vd); os.IsNotExist(err) {
				app.Logger().Error("Video dir not exists: ", vd)
				ctx.StatusCode(iris.StatusBadRequest)
				ctx.JSON(iris.Map{"status": iris.StatusBadRequest, "error": vd})
				return
			}

			list := make([]string, 0, 10)

			app.Logger().Info("list video dir: ", vd)
			err = filepath.Walk(vd, func(path string, info os.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}
				list = append(list, path[vdLen:])
				return nil
			})
			if err != nil {
				ctx.StatusCode(iris.StatusRequestedRangeNotSatisfiable)
				ctx.JSON(iris.Map{"error": "volume" + volumeId})
				return
			}

			ctx.JSON(list)
		})

		v1.Post("/file", func(ctx iris.Context) {
			var vf VideoFile
			err := ctx.ReadJSON(&vf)
			if err != nil {
				app.Logger().Error("file get ReadJSON error: ", err.Error())
				ctx.StatusCode(iris.StatusBadRequest)
				ctx.JSON(iris.Map{"status": iris.StatusBadRequest, "error": err.Error()})
				return
			}
			if vf.VolumeId == "" || vf.Path == "" {
				app.Logger().Error("file get not defined VolumeId or Path")
				ctx.StatusCode(iris.StatusBadRequest)
				ctx.JSON(iris.Map{"status": iris.StatusBadRequest, "error": "not defined VolumeId or Path"})
				return
			}

			fp, err := filepath.Abs(filepath.Join(conf.VideoDirsMap[vf.VolumeId], vf.Path))
			if err != nil {
				app.Logger().Error("Video file get full path error", vf.VolumeId)
				ctx.StatusCode(iris.StatusBadRequest)
				ctx.JSON(iris.Map{"status": iris.StatusBadRequest, "error": vf.VolumeId})
				return
			}

			if _, err := os.Stat(fp); os.IsNotExist(err) {
				app.Logger().Error("Video file not exists: ", fp)
				ctx.StatusCode(iris.StatusBadRequest)
				ctx.JSON(iris.Map{"status": iris.StatusBadRequest, "error": fp})
				return
			}

			app.Logger().Info("Send file: ", fp)
			ctx.SendFile(fp, filepath.Base(fp))
		})
	}

	app.Configure(iris.WithCharset("UTF-8"))
	app.Run(iris.TLS(conf.ServerAddr, conf.Cert, conf.Key))
}
