package main

import (
	"os"
	"time"
	"path/filepath"
	"io/ioutil"
	"crypto/rsa"

	"golang.org/x/crypto/bcrypt"

	// fork for windows compilation
	// "github.com/foomo/htpasswd"
	"github.com/sv99/htpasswd"

	"github.com/BurntSushi/toml"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/dgrijalva/jwt-go"
	jwtmiddleware "github.com/iris-contrib/middleware/jwt"
)

const VERSION = "0.1"

type Config struct {
	LogLevel   string
	ServerAddr string
	Cert       string
	Key        string
	VideoDirs  []string

	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
	passwords htpasswd.HashedPasswords
}

type Path struct {
	Path 	[]string `json:"path"`
}

type Token struct {
	Token string `json:"token"`
}

type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Response struct {
	Data string `json:"data"`
}

func validate(user *UserCredentials, conf *Config) bool {
	if hash, found := conf.passwords[user.Username]; found {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(user.Password))
		if err != nil {
			return false
		}
		return true
	}
	return false
}

func newLogFile() *os.File {
	filename, err := filepath.Abs(filepath.Join("log", "videodir.log"))
	if err != nil {
		panic(err.Error())
	}
	_ = os.Mkdir("log", 0755)
	// open an output file, this will append to the today's file if server restarted.
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err.Error())
	}

	return f
}

func main() {

	app := iris.New()
	app.Configure(iris.WithConfiguration(iris.Configuration{
		DisableStartupLog:                 false,
		Charset:                           "UTF-8",
	}))

	app.Logger().SetLevel("debug")
	app.Logger().AddOutput(newLogFile())

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
		LogLevel:   "info",
		ServerAddr: ":8443",
		Cert:       "server.crt",
		Key:        "server.key",
	}
	if _, err := toml.DecodeFile("./videodir.conf", &conf); err != nil {
		app.Logger().Warn("Config problems: " + err.Error())
	}
	// set LogLevel from config
	app.Logger().SetLevel(conf.LogLevel)

	// read private key
	app.Logger().Info("key: ", conf.Key)
	signBytes, err := ioutil.ReadFile(conf.Key)
	if err != nil {
		app.Logger().Fatal("read key file error: ", err.Error())
		return
	}
	conf.signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		app.Logger().Fatal("init private key error: ", err.Error())
		return
	}

	// read public key from certificate
	app.Logger().Info("cert: ", conf.Cert)
	certBytes, err := ioutil.ReadFile(conf.Cert)
	if err != nil {
		app.Logger().Fatal("read certificate file error: ", err.Error())
		return
	}
	conf.verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(certBytes)
	if err != nil {
		app.Logger().Fatal("init public key from certificate error: ", err.Error())
		return
	}

	// read .htpasswd
	conf.passwords, err = htpasswd.ParseHtpasswdFile(".htpasswd")
	if err != nil {
		app.Logger().Fatal("read .htpasswd error: ", err.Error())
		return
	}

	app.Favicon("./favicon.ico")

	app.RegisterView(iris.HTML("./", ".html"))
	app.Get("/", func(ctx iris.Context) {
		ctx.View("index.html")
	})

	app.Post("/login", func(ctx iris.Context) {
		var user UserCredentials
		err := ctx.ReadJSON(&user)
		if err != nil {
			app.Logger().Error("user credentials read error: ", err.Error())
			ctx.StatusCode(iris.StatusBadRequest)
			ctx.JSON(iris.Map{"status": iris.StatusBadRequest, "error": err.Error()})
			return
		}

		// validate username and password
		if !validate(&user, &conf) {
			ctx.StatusCode(iris.StatusUnauthorized)
			app.Logger().Error("user credentials check error: ", user.Username)
			ctx.JSON(iris.Map{"status": iris.StatusUnauthorized, "error": "Invalid credentials"})
			return
		}

		token := jwt.New(jwt.SigningMethodRS384)
		claims := make(jwt.MapClaims)
		claims["username"] = user.Username
		claims["exp"] = time.Now().Add(time.Hour * time.Duration(8)).Unix()
		claims["iat"] = time.Now().Unix()
		token.Claims = claims

		tokenString, err := token.SignedString(conf.signKey)
		if err != nil {
			ctx.StatusCode(iris.StatusInternalServerError)
			app.Logger().Error("Error while signing the token")
			ctx.JSON(iris.Map{"status": iris.StatusUnauthorized, "error": "Error while signing the token"})
			return
		}

		ctx.JSON(Token{tokenString})
	})

	//app.StaticWeb("/video1", *videoDir1)

	// v1 authenticated
	jwtHandler := jwtmiddleware.New(jwtmiddleware.Config{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return conf.verifyKey, nil
		},
		SigningMethod: jwt.SigningMethodRS384,
		//Debug: true,
	})

	// need JWT auth header
	v1 := app.Party("/api/v1", jwtHandler.Serve)
	{
		v1.Get("/version", func(ctx iris.Context) {
			ver := iris.Map{"version": VERSION}
			ctx.JSON(ver)
			ver = nil
		})

		v1.Get("/volumes", func(ctx iris.Context) {
			ctx.JSON(conf.VideoDirs)
		})

		v1.Post("/list", func(ctx iris.Context) {
			var vf Path
			err := ctx.ReadJSON(&vf)
			if err != nil {
				app.Logger().Error("file get ReadJSON error: ", err.Error())
				ctx.StatusCode(iris.StatusBadRequest)
				ctx.JSON(iris.Map{"status": iris.StatusBadRequest, "error": err.Error()})
				return
			}

			list := make([]string, 0, 10)

			for _, volume := range conf.VideoDirs {
				vd, err := filepath.Abs(filepath.Join(volume, filepath.Join(vf.Path...)))
				if err != nil {
					app.Logger().Error("Get full path error", vf.Path)
					ctx.StatusCode(iris.StatusBadRequest)
					ctx.JSON(iris.Map{"status": iris.StatusBadRequest, "error": vd})
					return
				}

				// this path not exists in current volume
				if _, err := os.Stat(vd); os.IsNotExist(err) {
					continue
				}

				files, err := ioutil.ReadDir(vd)
				if err != nil {
					app.Logger().Error("Read dir error: ", err.Error())
					ctx.StatusCode(iris.StatusBadRequest)
					ctx.JSON(iris.Map{"status": iris.StatusBadRequest, "error": err.Error()})
				}

				for _, f := range files {
					list = append(list, f.Name())
				}
			}

			ctx.JSON(list)
		})

		v1.Post("/file", func(ctx iris.Context) {
			var vf Path
			err := ctx.ReadJSON(&vf)
			if err != nil {
				app.Logger().Error("file get ReadJSON error: ", err.Error())
				ctx.StatusCode(iris.StatusBadRequest)
				ctx.JSON(iris.Map{"status": iris.StatusBadRequest, "error": err.Error()})
				return
			}

			// empty path not available
			if len(vf.Path) == 0 {
				app.Logger().Error("file path not specified")
				ctx.StatusCode(iris.StatusBadRequest)
				ctx.JSON(iris.Map{"status": iris.StatusBadRequest, "error": "file path not specified"})
				return
			}

			var fp = ""
			for _, volume := range conf.VideoDirs {

				fp, err = filepath.Abs(filepath.Join(volume, filepath.Join(vf.Path...)))
				if err != nil {
					app.Logger().Error("Video file get full path error " + err.Error())
					ctx.StatusCode(iris.StatusBadRequest)
					ctx.JSON(iris.Map{"status": iris.StatusBadRequest, "error": err.Error()})
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
	}

	app.Run(iris.TLS(conf.ServerAddr, conf.Cert, conf.Key))
}
