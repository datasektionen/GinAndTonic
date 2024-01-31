package jobs

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	// Set log output to the file
	log.SetOutput(os.Stdout)

	// Set log level
	log.SetLevel(logrus.InfoLevel)

	// Log as JSON for structured logging
	log.SetFormatter(&logrus.JSONFormatter{})
}
