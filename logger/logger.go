package logger

type Logger interface {
	Fatalf(string, ...interface{})
	Debugf(string, ...interface{})
	Errorf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Warningf(string, ...interface{})
	Debug(...interface{})
	Warn(...interface{})
	Warning(...interface{})
	Info(...interface{})
	Fatal(...interface{})
}
