// Copyright 2021 lack
// MIT License
//
// Copyright (c) 2020 Lack
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

package zap

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/natefinch/lumberjack"
	"github.com/vine-io/vine/lib/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLog struct {
	cfg    zap.Config
	logger *zap.Logger
	opts   logger.Options
	sync.RWMutex
	fields map[string]interface{}
}

// New builds a new logger based on options
func New(opts ...logger.Option) (*ZapLog, error) {
	// Default options
	options := logger.Options{
		Level:   logger.InfoLevel,
		Fields:  make(map[string]interface{}),
		Out:     os.Stderr,
		Context: context.Background(),
	}

	l := &ZapLog{opts: options}
	if err := l.Init(opts...); err != nil {
		return nil, err
	}

	return l, nil
}

func (l *ZapLog) Init(opts ...logger.Option) error {
	var err error

	for _, o := range opts {
		o(&l.opts)
	}

	zapConfig := zap.NewProductionConfig()
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	if zconfig, ok := l.opts.Context.Value(configKey{}).(zap.Config); ok {
		zapConfig = zconfig
	}

	if zcconfig, ok := l.opts.Context.Value(encoderConfigKey{}).(zapcore.EncoderConfig); ok {
		zapConfig.EncoderConfig = zcconfig
	}

	format := l.opts.Context.Value(encoderJSONKey{})

	zopts := make([]zap.Option, 0)
	skip, ok := l.opts.Context.Value(callerSkipKey{}).(int)
	if !ok || skip < 1 {
		skip = 1
	}
	zopts = append(zopts, zap.AddCallerSkip(skip))

	// Set log Level if not default
	zapConfig.Level = zap.NewAtomicLevel()
	if l.opts.Level != logger.InfoLevel {
		zapConfig.Level.SetLevel(loggerToZapLevel(l.opts.Level))
	}

	zopts = append(zopts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		cfg := zapConfig.EncoderConfig
		encoder := zapcore.NewConsoleEncoder(cfg)
		if format != nil {
			encoder = zapcore.NewJSONEncoder(cfg)
		}
		cc := zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), zapConfig.Level)
		return cc
	}))

	log, err := zapConfig.Build(zopts...)
	if err != nil {
		return err
	}

	// Adding seed fields if exist
	if l.opts.Fields != nil {
		data := []zap.Field{}
		for k, v := range l.opts.Fields {
			data = append(data, zap.Any(k, v))
		}
		log = log.With(data...)
	}

	// Adding namespace
	if namespace, ok := l.opts.Context.Value(namespaceKey{}).(string); ok {
		log = log.With(zap.Namespace(namespace))
	}

	if fw, ok := l.opts.Context.Value(FileWriter{}).(*lumberjack.Logger); ok {
		log = log.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			cfg := zapConfig.EncoderConfig
			encoder := zapcore.NewConsoleEncoder(cfg)
			if format != nil {
				encoder = zapcore.NewJSONEncoder(cfg)
			}
			cc := zapcore.NewCore(encoder, zapcore.AddSync(fw), zapConfig.Level)
			return cc
		}))
	}

	if writer, ok := l.opts.Context.Value(writerKey{}).(zapcore.WriteSyncer); ok {
		log = log.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			cfg := zapConfig.EncoderConfig
			encoder := zapcore.NewConsoleEncoder(cfg)
			if format != nil {
				encoder = zapcore.NewJSONEncoder(cfg)
			}
			cc := zapcore.NewCore(encoder, writer, zapConfig.Level)
			return zapcore.NewTee(core, cc)
		}))
	}

	// defer log.Sync() ??

	l.cfg = zapConfig
	l.logger = log
	l.fields = make(map[string]interface{})

	return nil
}

func (l *ZapLog) Fields(fields map[string]interface{}) logger.Logger {
	l.Lock()
	nfields := make(map[string]interface{}, len(l.fields))
	for k, v := range fields {
		nfields[k] = v
	}

	data := make([]zap.Field, 0, len(nfields))
	for k, v := range fields {
		data = append(data, zap.Any(k, v))
	}

	zl := &ZapLog{
		cfg:    l.cfg,
		logger: l.logger.With(data...),
		opts:   l.opts,
		fields: make(map[string]interface{}),
	}

	return zl
}

func (l *ZapLog) Error(err error) logger.Logger {
	return l.Fields(map[string]interface{}{"error": err})
}

func (l *ZapLog) Options() logger.Options {
	return l.opts
}

func (l *ZapLog) GetLogger() *zap.Logger {
	return l.logger
}

func (l *ZapLog) Log(level logger.Level, args ...interface{}) {
	l.RLock()
	data := make([]zap.Field, 0, len(l.fields))
	for k, v := range l.fields {
		data = append(data, zap.Any(k, v))
	}
	l.RUnlock()

	lvl := loggerToZapLevel(level)
	msg := fmt.Sprint(args...)
	switch lvl {
	case zap.DebugLevel:
		l.logger.Debug(msg, data...)
	case zap.InfoLevel:
		l.logger.Info(msg, data...)
	case zap.WarnLevel:
		l.logger.Warn(msg, data...)
	case zap.ErrorLevel:
		l.logger.Error(msg, data...)
	case zap.FatalLevel:
		l.logger.Fatal(msg, data...)
	}
}

func (l *ZapLog) Logf(level logger.Level, format string, args ...interface{}) {
	l.RLock()
	data := make([]zap.Field, 0, len(l.fields))
	for k, v := range l.fields {
		data = append(data, zap.Any(k, v))
	}
	l.RUnlock()

	lvl := loggerToZapLevel(level)
	msg := fmt.Sprintf(format, args...)
	switch lvl {
	case zap.DebugLevel:
		l.logger.Debug(msg, data...)
	case zap.InfoLevel:
		l.logger.Info(msg, data...)
	case zap.WarnLevel:
		l.logger.Warn(msg, data...)
	case zap.ErrorLevel:
		l.logger.Error(msg, data...)
	case zap.FatalLevel:
		l.logger.Fatal(msg, data...)
	}
}

func (l *ZapLog) Sync() error {
	return l.logger.Sync()
}

func (l *ZapLog) String() string {
	return "zap"
}
