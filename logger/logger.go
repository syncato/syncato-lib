// Copyright 2015 The Syncato Authors.  All rights reserved.
// Use of this source code is governed by a AGPL
// license that can be found in the LICENSE file.

// Package logger defines the logger used by the daemon and libraries to log information.
package logger

import (
	"os"

	"github.com/Sirupsen/logrus"
)

// Logger is responsible for log information to a target supported by the log implementation
type Logger struct {
	rid string         // the request id
	log *logrus.Logger // the log implementation
}

// NewLogger creates a logger instance with a custom log level
// The log level specifies from which level start logging.
// The possible values for the log level are: 0=panic,1=fatal,2=error,3=warning,4=info,5=debug
func NewLogger(rid string, level int) *Logger {
	logr := logrus.New()
	logr.Level = logrus.Level(level)
	log := Logger{rid, logr}
	return &log
}

func (l *Logger) Debug(msg interface{}, fields map[string]interface{}) {
	host, _ := os.Hostname()
	l.log.WithField("RID", l.rid).WithField("HOST", host).WithFields(fields).Debug(msg)
}
func (l *Logger) Info(msg interface{}, fields map[string]interface{}) {
	host, _ := os.Hostname()
	l.log.WithField("RID", l.rid).WithField("HOST", host).WithFields(fields).Info(msg)
}
func (l *Logger) Warn(msg interface{}, fields map[string]interface{}) {
	host, _ := os.Hostname()
	l.log.WithField("RID", l.rid).WithField("HOST", host).WithFields(fields).Warn(msg)
}
func (l *Logger) Error(msg interface{}, fields map[string]interface{}) {
	host, _ := os.Hostname()
	l.log.WithField("RID", l.rid).WithField("HOST", host).WithFields(fields).Error(msg)
}
func (l *Logger) Fatal(msg interface{}, fields map[string]interface{}) {
	host, _ := os.Hostname()
	l.log.WithField("RID", l.rid).WithField("HOST", host).WithFields(fields).Fatal(msg)
}
func (l *Logger) Panic(msg interface{}, fields map[string]interface{}) {
	host, _ := os.Hostname()
	l.log.WithField("RID", l.rid).WithField("HOST", host).WithFields(fields).Panic(msg)
}
