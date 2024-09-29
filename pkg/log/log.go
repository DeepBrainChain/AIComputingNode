package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/ipfs/go-log/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger = log.Logger("AIComputingNode")

func InitLogging(levelString string, logFile string, logOutput string) error {
	logLevel, err := log.LevelFromString(levelString)
	if err != nil {
		return err
	}
	log.SetAllLoggers(logLevel)
	log.SetLogLevel("AIComputingNode", levelString)

	// multiWriteSyncer := make([]zapcore.WriteSyncer, 0)
	// outputOptions := strings.Split(logOutput, "+")
	// for _, opt := range outputOptions {
	// 	switch opt {
	// 	case "stdout":
	// 		multiWriteSyncer = append(multiWriteSyncer, zapcore.AddSync(os.Stdout))
	// 	case "stderr":
	// 		multiWriteSyncer = append(multiWriteSyncer, zapcore.AddSync(os.Stderr))
	// 	case "file":
	// 		if logFile != "" {
	// 			multiWriteSyncer = append(
	// 				multiWriteSyncer,
	// 				zapcore.AddSync(&lumberjack.Logger{
	// 					Filename:   logFile,
	// 					MaxSize:    10,   // Maximum file size (MB)
	// 					MaxBackups: 30,   // Keep up to 30 backups
	// 					MaxAge:     28,   // Maximum number of days to save files
	// 					Compress:   true, // Whether to disable compression for old files
	// 				}),
	// 			)
	// 		}
	// 	}
	// }
	// encCfg := zap.NewProductionEncoderConfig()
	// encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	// encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	// // encCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	// encoder := zapcore.NewConsoleEncoder(encCfg)
	// core := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(multiWriteSyncer...), zap.InfoLevel)
	// log.SetPrimaryCore(core)
	// return nil

	var outputStderr bool = false
	var outputStdout bool = false
	var outputFile bool = false
	outputOptions := strings.Split(logOutput, "+")
	for _, opt := range outputOptions {
		switch opt {
		case "stdout":
			outputStdout = true
			continue
		case "stderr":
			outputStderr = true
			continue
		case "file":
			outputFile = true
			continue
		}
	}

	// only console log
	if logFile == "" || !outputFile {
		os.Setenv("GOLOG_OUTPUT", logOutput)
		return nil
	}

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	// only file log
	if logFile != "" && !outputStderr && !outputStdout && outputFile {
		encCfg.EncodeLevel = zapcore.CapitalLevelEncoder
		encoder := zapcore.NewConsoleEncoder(encCfg)

		logWriter := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    10,   // Maximum file size (MB)
			MaxBackups: 30,   // Keep up to 30 backups
			MaxAge:     28,   // Maximum number of days to save files
			Compress:   true, // Whether to disable compression for old files
		}

		zapCore := zapcore.NewCore(encoder, zapcore.AddSync(logWriter), zap.InfoLevel)
		log.SetPrimaryCore(zapCore)
		return nil
	}

	// both console log and file log
	outputPaths := []string{}
	if outputStderr {
		outputPaths = append(outputPaths, "stderr")
	}
	if outputStdout {
		outputPaths = append(outputPaths, "stdout")
	}
	ws, _, err := zap.Open(outputPaths...)
	if err != nil {
		return fmt.Errorf("unable to open logging output")
	}
	encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(encCfg)
	consoleCore := zapcore.NewCore(consoleEncoder, ws, zap.InfoLevel)

	encCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	fileEncoder := zapcore.NewConsoleEncoder(encCfg)

	logWriter := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10,   // Maximum file size (MB)
		MaxBackups: 30,   // Keep up to 30 backups
		MaxAge:     28,   // Maximum number of days to save files
		Compress:   true, // Whether to disable compression for old files
	}

	fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(logWriter), zap.InfoLevel)

	tee := zap.New(zapcore.NewTee(
		consoleCore,
		fileCore,
	), zap.AddCaller())
	log.SetPrimaryCore(tee.Core())

	return nil
}
