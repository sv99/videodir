package videodir

import (
	"crypto/tls"
	"log"
	"path/filepath"

	"github.com/dgrijalva/jwt-go"
	"github.com/foomo/htpasswd"
	"github.com/gofiber/fiber"
	"github.com/gofiber/helmet"
	jwtware "github.com/gofiber/jwt"
	"github.com/gofiber/recover"
	"github.com/rs/zerolog"
)

const (
	VERSION  = "0.4"
	HTPASSWD = "htpasswd"
	CONF_FILE = "videodir.conf"
)

type AppServer struct {
	App    *fiber.App
	Logger *zerolog.Logger
	Config *Config

	WorkDir   string
	Passwords htpasswd.HashedPasswords
}

func (srv *AppServer) Error(c *fiber.Ctx, status int, message string) {
	srv.Logger.Info().Msg(message)
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

func NewApp(confPath string, zeroLogger *zerolog.Logger) *AppServer {
	conf := DefaultConfig()
	err := conf.TOML(confPath)
	if err != nil {
		log.Panicf("Config load problems: %v", err)
	}

	app := fiber.New()

	srv := AppServer{
		Config:  &conf,
		App:     app,
		WorkDir: filepath.Dir(confPath),
		Logger: zeroLogger,
	}

	// set global log level
	if srv.Config.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// dump config
	srv.Logger.Debug().Msgf("Config: %s", confPath)
	srv.Logger.Debug().Msgf("Cert: %s", srv.Config.Cert)
	srv.Logger.Debug().Msgf("Key: %s", srv.Config.Key)
	srv.Logger.Debug().Msgf("Videodirs: %s", srv.Config.VideoDirs)

	// read .htpasswd
	srv.Passwords, err = htpasswd.ParseHtpasswdFile(GetHtPasswdPath(srv.WorkDir))
	if err != nil {
		srv.Logger.Fatal().Msg("read htpasswd error: " + err.Error())
	}

	//app.Use(logger.New(logger.Config{
	//	//Output: newLogFile(),
	//	TimeFormat: "2006-01-02 15:04:05",
	//}))
	app.Use(NewLoggerMiddleware(zeroLogger))

	app.Use(recover.New(recover.Config{
		Handler: func(c *fiber.Ctx, err error) {
			c.Status(500)
			_ = c.JSON(fiber.Map{"Message": err.Error()})
		},
	}))
	app.Use(helmet.New())

	// index - login not require
	app.Get("/", func(ctx *fiber.Ctx) {
		//_ = ctx.SendFile("./index.html")
		ctx.Type("html", "utf-8")
		ctx.Send(_indexHtml)
	})
	app.Get("/favicon.ico", func(ctx *fiber.Ctx) {
		//_ = ctx.SendFile("./favicon.ico")
		ctx.Type("ico")
		ctx.Send(_faviconIco)
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
	srv.Logger.Info().Msg("Start server")
	pathCert := filepath.Join(srv.WorkDir, srv.Config.Cert)
	pathKey := filepath.Join(srv.WorkDir, srv.Config.Key)
	cert, err := tls.LoadX509KeyPair(pathCert, pathKey)
	if err != nil {
		srv.Logger.Fatal().Msgf("Error create tls certificate: %v", err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	err = srv.App.Listen(srv.Config.ServerAddr, config)
	if err != nil {
		srv.Logger.Fatal().Msgf("Listen error: %v", err)
	}
	srv.Logger.Info().Msg("Server stopped")
}

func (srv *AppServer) Shutdown() {
	err := srv.App.Shutdown()
	if err != nil {
		srv.Logger.Fatal().Msgf("Shutdown error %v", err)
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
