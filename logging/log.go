// Copyright 2020 lesismal. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package logging

import (
	"fmt"
	"os"
	"time"
)

var (
	// TimeFormat is used to format time parameters.
	TimeFormat = "2006/01/02 15:04:05.000"

	// Output is used to receive log output.
	Output = os.Stdout

	// DefaultLogger is the default logger and is used by arpc.
	DefaultLogger Logger = &logger{level: LevelInfo}
)

const (
	// LevelAll enables all logs.
	LevelAll = iota
	// LevelDebug logs are usually disabled in production.
	LevelDebug
	// LevelInfo is the default logging priority.
	LevelInfo
	// LevelWarn .
	LevelWarn
	// LevelError .
	LevelError
	// LevelNone disables all logs.
	LevelNone
)

// Logger defines log interface.
type Logger interface {
	SetLevel(lvl int)
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

// SetLogger sets default logger.
func SetLogger(l Logger) {
	DefaultLogger = l
}

// SetLevel sets default logger's priority.
func SetLevel(lvl int) {
	switch lvl {
	case LevelAll, LevelDebug, LevelInfo, LevelWarn, LevelError, LevelNone:
		DefaultLogger.SetLevel(lvl)
	default:
		fmt.Fprintf(Output, "invalid log level: %v", lvl)
	}
}

// logger implements Logger and is used in arpc by default.
type logger struct {
	level int
}

// SetLevel sets logs priority.
func (l *logger) SetLevel(lvl int) {
	switch lvl {
	case LevelAll, LevelDebug, LevelInfo, LevelWarn, LevelError, LevelNone:
		l.level = lvl
	default:
		fmt.Fprintf(Output, "invalid log level: %v", lvl)
	}
}

// Debug uses fmt.Printf to log a message at LevelDebug.
func (l *logger) Debugf(format string, v ...interface{}) {
	if LevelDebug >= l.level {
		fmt.Fprintf(Output, time.Now().Format(TimeFormat)+" [DBG] "+format+"\n", v...)
	}
}

// Info uses fmt.Printf to log a message at LevelInfo.
func (l *logger) Infof(format string, v ...interface{}) {
	if LevelInfo >= l.level {
		fmt.Fprintf(Output, time.Now().Format(TimeFormat)+" [INF] "+format+"\n", v...)
	}
}

// Warn uses fmt.Printf to log a message at LevelWarn.
func (l *logger) Warnf(format string, v ...interface{}) {
	if LevelWarn >= l.level {
		fmt.Fprintf(Output, time.Now().Format(TimeFormat)+" [WRN] "+format+"\n", v...)
	}
}

// Error uses fmt.Printf to log a message at LevelError.
func (l *logger) Errorf(format string, v ...interface{}) {
	if LevelError >= l.level {
		fmt.Fprintf(Output, time.Now().Format(TimeFormat)+" [ERR] "+format+"\n", v...)
	}
}

// Debugf uses DefaultLogger to log a message at LevelDebug.
func Debugf(format string, v ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Debugf(format, v...)
	}
}

// Infof uses DefaultLogger to log a message at LevelInfo.
func Infof(format string, v ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Infof(format, v...)
	}
}

// Warnf uses DefaultLogger to log a message at LevelWarn.
func Warnf(format string, v ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Warnf(format, v...)
	}
}

// Errorf uses DefaultLogger to log a message at LevelError.
func Errorf(format string, v ...interface{}) {
	if DefaultLogger != nil {
		DefaultLogger.Errorf(format, v...)
	}
}
