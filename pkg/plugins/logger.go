package plugins

import log "github.com/sirupsen/logrus"

type PluginLogger struct{}

func (p PluginLogger) Trace(a ...interface{}) {
	log.Trace(a...)
}

func (p PluginLogger) Tracef(format string, a ...interface{}) {
	log.Tracef(format, a...)
}

func (p PluginLogger) Debug(a ...interface{}) {
	log.Debug(a...)
}

func (p PluginLogger) Debugf(format string, a ...interface{}) {
	log.Debugf(format, a...)
}

func (p PluginLogger) Info(a ...interface{}) {
	log.Info(a...)
}

func (p PluginLogger) Infof(format string, a ...interface{}) {
	log.Infof(format, a...)
}

func (p PluginLogger) Warn(a ...interface{}) {
	log.Warn(a...)
}

func (p PluginLogger) Warnf(format string, a ...interface{}) {
	log.Warnf(format, a...)
}

func (p PluginLogger) Error(a ...interface{}) {
	log.Error(a...)
}

func (p PluginLogger) Errorf(format string, a ...interface{}) {
	log.Errorf(format, a...)
}

func (p PluginLogger) Fatal(a ...interface{}) {
	log.Fatal(a...)
}

func (p PluginLogger) Fatalf(format string, a ...interface{}) {
	log.Fatalf(format, a...)
}

func (p PluginLogger) Panic(a ...interface{}) {
	log.Panic(a...)
}

func (p PluginLogger) Panicf(format string, a ...interface{}) {
	log.Panicf(format, a...)
}
