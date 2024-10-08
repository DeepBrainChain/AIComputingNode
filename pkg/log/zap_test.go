package log

import (
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestZapSugaredLogger(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	logger.Info(
		"failed to fetch URL",
		// 字段是强类型，不是松散类型
		zap.String("url", "https://example.com"),
		zap.Int("attempt", 3),
		zap.Duration("backoff", time.Second),
	)

	sugar := logger.Sugar()
	sugar.Infow(
		"failed to fetch URL",
		// 字段是松散类型，不是强类型
		"url", "https://example.com",
		"attempt", 3,
		"backoff", time.Second,
	)
	sugar.Infow(
		"failed to fetch URL",
		zap.String("url", "https://example.com"),
		zap.Int("attempt", 3),
		zap.Duration("backoff", time.Second),
	)
	fields := []zapcore.Field{
		zap.String("url", "https://example.com"),
		zap.Int("attempt", 3),
		zap.Duration("backoff", time.Second),
	}
	sugar.Infow("failed to fetch URL", fields)
	sugar.Desugar().Info("failed to fetch URL", fields...)
}

func TestZapProduction(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	logger.Debug("this is debug message")
	logger.Info("this is info message")
	logger.Info("this is info message with fileds",
		zap.Int("age", 37),
		zap.String("agender", "man"),
	)
	logger.Warn("this is warn message")
	logger.Error("this is error message")
}

func TestZapExample(t *testing.T) {
	logger := zap.NewExample()
	defer logger.Sync()
	logger.Debug("this is debug message")
	logger.Info("this is info message")
	logger.Info("this is info message with fileds",
		zap.Int("age", 37),
		zap.String("agender", "man"),
	)
	logger.Warn("this is warn message")
	logger.Error("this is error message")
}

func TestZapDevelopment(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	logger.Debug("this is debug message")
	logger.Info("this is info message")
	logger.Info("this is info message with fileds",
		zap.Int("age", 37),
		zap.String("agender", "man"),
	)
	logger.Warn("this is warn message")
	logger.Error("this is error message")
}

func TestZapCustomLogger(t *testing.T) {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	// encoder := zapcore.NewJSONEncoder(config)
	encoder := zapcore.NewConsoleEncoder(config)
	// file, _ := os.Create("./test.log")
	// writeSyncer := zapcore.AddSync(file)
	// 利用 io.MultiWriter 支持文件和终端两个输出目标
	// file, _ := os.Create("./test.log")
	// ws := io.MultiWriter(file, os.Stdout)
	// writeSyncer := zapcore.AddSync(ws)

	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zap.InfoLevel)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	defer logger.Sync()
	logger.Debug("this is debug message")
	logger.Info("this is info message")
	logger.Info("this is info message with fileds",
		zap.Int("age", 37),
		zap.String("agender", "man"),
	)
	logger.Warn("this is warn message")
	logger.Error("this is error message")
}
