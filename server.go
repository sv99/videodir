package videodir

import (
	"crypto/tls"
	"github.com/dgrijalva/jwt-go"
	"github.com/foomo/htpasswd"
	"github.com/gofiber/fiber"
	"github.com/gofiber/helmet"
	jwtware "github.com/gofiber/jwt"
	"github.com/gofiber/logger"
	"github.com/gofiber/recover"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

const (
	VERSION  = "0.3"
	HTPASSWD = "htpasswd"
)

type AppServer struct {
	App    *fiber.App
	Logger *logrus.Logger
	Config *Config

	passwords htpasswd.HashedPasswords
}

func (srv *AppServer) Error(c *fiber.Ctx, status int, message string) {
	srv.Logger.Info(message)
	c.SendStatus(status)
	_ = c.JSON(fiber.Map{"status": status, "error": message})
}

func (srv *AppServer) NotFound(c *fiber.Ctx) {
	srv.Error(
		c,
		fiber.StatusNotFound,
		"Sorry, but the page you were looking for could not be found.",
	)
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

func NewApp(conf *Config) *AppServer {
	app := fiber.New()

	srv := AppServer{
		Config: conf,
		App:    app,
	}

	// init logger
	srv.InitLogger()

	// read .htpasswd
	var err error
	srv.passwords, err = htpasswd.ParseHtpasswdFile(HTPASSWD)
	if err != nil {
		srv.Logger.Fatal("read htpasswd error: ", err.Error())
	}

	app.Use(logger.New(logger.Config{
		//Output: newLogFile(),
		TimeFormat: "2006-01-02 15:04:05",
	}))
	app.Use(recover.New(recover.Config{
		Handler: func(c *fiber.Ctx, err error) {
			c.Status(500)
			_ = c.JSON(fiber.Map{"Message": err.Error()})
		},
	}))
	app.Use(helmet.New())

	// index - login not require
	app.Get("/", func(ctx *fiber.Ctx) {
		_ = ctx.SendFile("./index.html")
	})
	app.Get("/favicon.ico", func(ctx *fiber.Ctx) {
		_ = ctx.SendFile("./favicon.ico")
	})
	// Login route
	app.Post("/login", srv.Login)

	//app.Use(dumpHeaders)
	// JWT Middleware
	app.Use(jwtware.New(jwtware.Config{
		SigningKey:     []byte(srv.Config.JwtSecret),
		SigningMethod:  "HS384",
		ContextKey:     "token",
		SuccessHandler: jwtSuccessHandler,
		ErrorHandler:   jwtErrorHandler,
	}))

	// API
	v1 := app.Group("/api/v1")

	v1.Get("/version", func(ctx *fiber.Ctx) {
		_ = ctx.JSON(fiber.Map{"version": VERSION})
	})
	v1.Get("/volumes", func(ctx *fiber.Ctx) {
		_ = ctx.JSON(conf.VideoDirs)
	})

	v1.Post("/list", srv.ListPath)
	v1.Post("/file", srv.PostFile)
	v1.Post("/filesize", srv.PostFileSize)
	v1.Post("/remove", srv.RemoveFile)

	app.Use(srv.NotFound)
	srv.App = app
	return &srv
}

func (srv *AppServer) Serve() {
	// Create tls certificate
	cert, err := tls.LoadX509KeyPair(srv.Config.Cert, srv.Config.Key)
	if err != nil {
		srv.Logger.Fatalf("Error create tls certificate: %s", err.Error())
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	err = srv.App.Listen(srv.Config.ServerAddr, config)
	if err != nil {
		srv.Logger.Fatalf("Listen error: %s", err.Error())
	}
}

func (srv *AppServer) Shutdown() {
	err := srv.App.Shutdown()
	if err != nil {
		srv.Logger.Fatal("Shutdown error", err.Error())
	}
}

func jwtSuccessHandler(c *fiber.Ctx) {
	token := c.Locals("token").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	c.Locals("username", claims["username"].(string))
	c.Next()
}

func jwtErrorHandler(c *fiber.Ctx, _ error) {
	c.Status(fiber.StatusUnauthorized)
	c.SendString("Invalid, missing, malformed  or expired JWT")
}
