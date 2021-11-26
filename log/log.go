package log

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// Logger ...
var Logger *log.Logger

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
	Logger = log.New()
}
