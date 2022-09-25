package interfaces

type LoggerInterface interface {
	Debug(a ...interface{})
	Debugf(format string, a ...interface{})
	Info(a ...interface{})
	Infof(format string, a ...interface{})
	Warn(a ...interface{})
	Warnf(format string, a ...interface{})
	Error(a ...interface{})
	Errorf(format string, a ...interface{})
	Fatal(a ...interface{})
	Fatalf(format string, a ...interface{})
	Panic(a ...interface{})
	Panicf(format string, a ...interface{})
	Sync() error
}
