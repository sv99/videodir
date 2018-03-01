package main

import (
	"errors"
	"os"
	"time"
	"path/filepath"
	"crypto/rsa"

	"golang.org/x/crypto/bcrypt"

	// fork for windows compilation
	// "github.com/foomo/htpasswd"
	"github.com/sv99/htpasswd"

	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/dgrijalva/jwt-go"
	jwtmiddleware "github.com/iris-contrib/middleware/jwt"
)

const (
	VERSION = "0.2"
	HTPASSWD = "htpasswd"
)

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
	Path []string `json:"path"`
}

type Token struct {
	Token string `json:"token"`
}

type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
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

func sendJsonError(app *iris.Application, ctx iris.Context, status int, message string) {
	app.Logger().Error(message)
	ctx.StatusCode(status)
	ctx.JSON(iris.Map{"status": status, "error": message})
}

func readJsonPath(app *iris.Application, ctx iris.Context) (Path, error) {
	var vf Path
	err := ctx.ReadJSON(&vf)
	if err != nil {
		sendJsonError(app, ctx, iris.StatusBadRequest,
			"filesize ReadJSON error: "+err.Error())
		return vf, err
	}

	// empty path not available
	if len(vf.Path) == 0 {
		sendJsonError(app, ctx, iris.StatusBadRequest,
			"file path not specified")
		return vf, errors.New("file path not specified")
	}
	return vf, nil
}

func main() {

	app := iris.New()
	app.Configure(iris.WithConfiguration(iris.Configuration{
		DisableStartupLog: false,
		Charset:           "UTF-8",
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
	conf.Init(app)

	app.Favicon("./favicon.ico")

	app.RegisterView(iris.HTML("./", ".html"))
	app.Get("/", func(ctx iris.Context) {
		ctx.View("index.html")
	})

	app.Post("/login", func(ctx iris.Context) {
		var user UserCredentials
		err := ctx.ReadJSON(&user)
		if err != nil {
			sendJsonError(app, ctx, iris.StatusBadRequest,
				"user credentials read error: "+err.Error())
			return
		}

		// validate username and password
		if !validate(&user, &conf) {
			sendJsonError(app, ctx, iris.StatusUnauthorized,
				"invalid user")
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
			sendJsonError(app, ctx, iris.StatusInternalServerError,
				"Error while signing the token: "+err.Error())
			return
		}

		ctx.JSON(Token{tokenString})
	})

	// api v1 authentication handler
	jwtHandler := jwtmiddleware.New(jwtmiddleware.Config{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return conf.verifyKey, nil
		},
		SigningMethod: jwt.SigningMethodRS384,
		//Debug: true,
	})

	// need JWT auth header
	v1 := app.Party("/api/v1", jwtHandler.Serve)

	InitApi(app, v1, &conf)

	cliApp := InitCli(app, &conf)
	cliApp.WithAction(func(args []string, options map[string]string) int {
		app.Run(iris.TLS(conf.ServerAddr, conf.Cert, conf.Key))
		return 0
	})

	os.Exit(cliApp.Run(os.Args, os.Stdout))
}
