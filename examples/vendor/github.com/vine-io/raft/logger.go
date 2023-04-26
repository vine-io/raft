// MIT License
//
// Copyright (c) 2023 Lack
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
