package videodir

import (
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

func NewLogger(workDir string, console bool) (*zerolog.Logger, error) {
	// log to file for all
	filename := filepath.Join(workDir, "log", "videodir.log")
	_ = os.Mkdir(filepath.Join(workDir, "log"), 0755)
	// open an output file, this will append to the today's file if server restarted.
	logfile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	fileWriter := zerolog.ConsoleWriter{
		Out:     logfile,
		NoColor: true,
		//TimeFormat: time.RFC3339,
		TimeFormat: "2006-01-02 15:04:05",
	}
	// service not log to file if exists console out!!!
	if console {
		consoleWriter := zerolog.ConsoleWriter{
			Out:     os.Stdout,
			NoColor: true,
			//TimeFormat: time.RFC3339,
			TimeFormat: "2006-01-02 15:04:05",
		}
		multi := zerolog.MultiLevelWriter(consoleWriter, fileWriter)
		logger := zerolog.New(multi).With().Timestamp().Logger()
		logger.Info().Msg("multi logger")
		return &logger, nil
	} else {
		logger := zerolog.New(fileWriter).With().Timestamp().Logger()
		logger.Info().Msg("file logger")
		return &logger, nil
	}
}

func NewLoggerMiddleware(logger *zerolog.Logger) fiber.Handler {
	// default format "${time} ${method} ${path} - ${ip} - ${status} - ${latency}\n"
	// Middleware function
	return func(c *fiber.Ctx) error {
		start := time.Now()
		// handle request
		err := c.Next()
		// build log
		stop := time.Now()

		logger.Info().Msgf("%s %s - %s - %d - %s",
			c.Method(), c.Path(), c.IP(), c.Response().StatusCode(),
			stop.Sub(start).String())
		return err
	}
}
