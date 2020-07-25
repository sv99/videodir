package videodir

import (
	"github.com/sirupsen/logrus"
	"os"
)

func (srv *AppServer) InitLogger() {

	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{})
	log.SetOutput(os.Stdout)

	if srv.Config.Debug {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.WarnLevel)
	}
	srv.Logger = log
}
