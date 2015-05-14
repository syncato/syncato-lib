package logger

import (
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
	"net/http"
	"os"
)

type Fields map[string]interface{}

type Logger struct {
	rid  string
	log  *logrus.Logger
	comp string
}

func NewLogger(rid string, level int) *Logger {
	logr := logrus.New()
	logr.Level = logrus.Level(level)
	log := Logger{rid, logr, ""}
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

func GetLoggerFromReq(r *http.Request) (log *Logger) {
	return context.Get(r, "log").(*Logger)
}
