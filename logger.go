package raft

import (
	"go.etcd.io/etcd/raft/v3"
	"go.uber.org/zap"
)

var _ raft.Logger = (*logger)(nil)

type logger struct {
	base *zap.Logger
}

func newLogger(base *zap.Logger) *logger {
	return &logger{base: base}
}

func (l *logger) Debug(v ...interface{}) {
	l.Log().Debug(v...)
}

func (l *logger) Debugf(format string, v ...interface{}) {
	l.Log().Debugf(format, v...)
}

func (l *logger) Error(v ...interface{}) {
	l.Log().Error(v...)
}

func (l *logger) Errorf(format string, v ...interface{}) {
	l.Log().Errorf(format, v...)
}

func (l *logger) Info(v ...interface{}) {
	l.Log().Info(v...)
}

func (l *logger) Infof(format string, v ...interface{}) {
	l.Log().Infof(format, v...)
}

func (l *logger) Warning(v ...interface{}) {
	l.Log().Warn(v...)
}

func (l *logger) Warningf(format string, v ...interface{}) {
	l.Log().Warnf(format, v...)
}

func (l *logger) Fatal(v ...interface{}) {
	l.Log().Fatal(v...)
}

func (l *logger) Fatalf(format string, v ...interface{}) {
	l.Log().Fatalf(format, v...)
}

func (l *logger) Panic(v ...interface{}) {
	l.Log().Panic(v...)
}

func (l *logger) Panicf(format string, v ...interface{}) {
	l.Log().Panicf(format, v...)
}

func (l *logger) Log() *zap.SugaredLogger {
	return l.base.With(zap.String("ns", "raft")).Sugar()
}
