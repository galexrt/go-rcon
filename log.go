package rcon

import (
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.Out = ioutil.Discard
}

// SetLog set sirupsen logger
func SetLog(l *logrus.Logger) {
	log = l
}
