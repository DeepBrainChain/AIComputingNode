package log

import (
	"github.com/ipfs/go-log/v2"
)

var Logger = log.Logger("AIComputingNode")

func SetLogLevel(levelString string) error {
	logLevel, err := log.LevelFromString(levelString)
	if err != nil {
		return err
	}
	log.SetAllLoggers(logLevel)
	log.SetLogLevel("AIComputingNode", levelString)
	return nil
}
