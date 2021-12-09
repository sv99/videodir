package videodir

import (
	"log"
	"os"
	"path/filepath"

	"github.com/foomo/htpasswd"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/helmet/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/zerolog"
)

const (
	VERSION   = "0.5"
	HTPASSWD  = "htpasswd"
	CONF_FILE = "videodir.conf"
)

type AppServer struct {
	App    *fiber.App
	Logger *zerolog.Logger
	Config *Config

	WorkDir   string
	Passwords htpasswd.HashedPasswords
}

func (srv *AppServer) Error(c *fiber.Ctx, status int, message string) error {
	srv.Logger.Info().Msg(message)
	_ = c.SendStatus(status)
	return c.JSON(fiber.Map{"status": status, "error": message})
}

func (srv *AppServer) NotFound(c *fiber.Ctx) error {
	return srv.Error(
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
		Logger:  zeroLogger,
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

	// check is htpasswd exists
	htpass := GetHtPasswdPath(srv.WorkDir)
	if _, err := os.Stat(htpass); os.IsNotExist(err) {
		srv.Logger.Debug().Msg("htpasswd does not exist")
		file, err := os.Create(htpass)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
	}
	// read htpasswd
	srv.Passwords, err = htpasswd.ParseHtpasswdFile(htpass)
	if err != nil {
		srv.Logger.Fatal().Msg("read htpasswd error: " + err.Error())
	}

	//app.Use(logger.New(logger.Config{
	//	//Output: newLogFile(),
	//	TimeFormat: "2006-01-02 15:04:05",
	//}))
	app.Use(NewLoggerMiddleware(zeroLogger))

	app.Use(recover.New())
	app.Use(helmet.New())

	// index - login not require
	app.Get("/", func(ctx *fiber.Ctx) error {
		//_ = ctx.SendFile("./index.html")
		ctx.Type("html", "utf-8")
		return ctx.Send(_indexHtml)
	})
	app.Get("/favicon.ico", func(ctx *fiber.Ctx) error {
		//_ = ctx.SendFile("./favicon.ico")
		ctx.Type("ico")
		return ctx.Send(_faviconIco)
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

	v1.Get("/version", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{"version": VERSION})
	})
	v1.Get("/volumes", func(ctx *fiber.Ctx) error {
		return ctx.JSON(conf.VideoDirs)
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

	err := srv.App.ListenTLS(srv.Config.ServerAddr, pathCert, pathKey)
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

func jwtSuccessHandler(c *fiber.Ctx) error {
	token := c.Locals("token").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	c.Locals("username", claims["username"].(string))
	return c.Next()
}

func jwtErrorHandler(c *fiber.Ctx, err error) error {
	if err.Error() == "Missing or malformed JWT" {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"status": "error", "message": "Missing or malformed JWT", "data": nil})
		// SendString("Missing or malformed JWT")
	}
	return c.Status(fiber.StatusUnauthorized).
		JSON(fiber.Map{"status": "error", "message": "Invalid or expired JWT", "data": nil})
	// SendString("Invalid, missing, malformed  or expired JWT")
}
