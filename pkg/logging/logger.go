package logging

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap/zapcore"
)
import "go.uber.org/zap"

type Logger struct {
	logger *zap.Logger
}

func (p *Logger) log(loglevel zapcore.Level, a ...interface{}) {
	p.logger.Log(loglevel, fmt.Sprintln(a...))
}

func (p *Logger) Debug(a ...interface{}) {
	p.log(zapcore.DebugLevel, a...)
}

func (p *Logger) Debugf(format string, a ...interface{}) {
	p.log(zapcore.DebugLevel, fmt.Sprintf(format, a...))
}

func (p *Logger) Info(a ...interface{}) {
	p.log(zapcore.InfoLevel, a...)
}

func (p *Logger) Infof(format string, a ...interface{}) {
	p.log(zapcore.InfoLevel, fmt.Sprintf(format, a...))
}

func (p *Logger) Warn(a ...interface{}) {
	p.log(zapcore.WarnLevel, a...)
}

func (p *Logger) Warnf(format string, a ...interface{}) {
	p.log(zapcore.WarnLevel, fmt.Sprintf(format, a...))
}

func (p *Logger) Error(a ...interface{}) {
	p.log(zapcore.ErrorLevel, a...)
}

func (p *Logger) Errorf(format string, a ...interface{}) {
	p.log(zapcore.ErrorLevel, fmt.Sprintf(format, a...))
}

func (p *Logger) Fatal(a ...interface{}) {
	p.log(zapcore.FatalLevel, a...)
}

func (p *Logger) Fatalf(format string, a ...interface{}) {
	p.log(zapcore.FatalLevel, fmt.Sprintf(format, a...))
}

func (p *Logger) Panic(a ...interface{}) {
	p.log(zapcore.PanicLevel, a...)
}

func (p *Logger) Panicf(format string, a ...interface{}) {
	p.log(zapcore.PanicLevel, fmt.Sprintf(format, a...))
}

func (p *Logger) Sync() error {
	return p.logger.Sync()
}

func NewLogger(level string, encoding string) *Logger {
	rawJSON := []byte(`{
	  "level": "` + level + `",
	  "encoding": "` + encoding + `",
	  "outputPaths": ["stdout"],
	  "errorOutputPaths": ["stderr"],
	  "encoderConfig": {
	    "messageKey": "message",
	    "levelKey": "level",
	    "levelEncoder": "lowercase"
	  }
	}`)

	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	logger := zap.Must(cfg.Build())

	return &Logger{
		logger: logger,
	}
}
