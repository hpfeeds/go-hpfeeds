package hpfeeds

import (
	"fmt"
)

// Logger is a function that will take a Printf style input string and argument
type Logger func(...interface{})

// SetDebugLogger sets the broker's logger output for Debug level logs
func (b *Broker) SetDebugLogger(logger Logger) {
	b.debugLogger = logger
}

func (b *Broker) logDebug(args ...interface{}) {
	if b.debugLogger != nil {
		b.debugLogger(args...)
	}
}

func (b *Broker) logDebugf(format string, args ...interface{}) {
	if b.infoLogger != nil {
		out := fmt.Sprintf(format, args...)
		b.infoLogger(out)
	}
}

// SetErrorLogger sets the broker's logger output for Error level logs
func (b *Broker) SetErrorLogger(logger Logger) {
	b.errorLogger = logger
}

func (b *Broker) logError(args ...interface{}) {
	if b.errorLogger != nil {
		b.errorLogger(args...)
	}
}

func (b *Broker) logErrorf(format string, args ...interface{}) {
	if b.infoLogger != nil {
		out := fmt.Sprintf(format, args...)
		b.infoLogger(out)
	}
}

// SetInfoLogger sets the broker's logger output for Info level logs
func (b *Broker) SetInfoLogger(logger Logger) {
	b.infoLogger = logger
}

func (b *Broker) logInfo(args ...interface{}) {
	if b.infoLogger != nil {
		b.infoLogger(args...)
	}
}

func (b *Broker) logInfof(format string, args ...interface{}) {
	if b.infoLogger != nil {
		out := fmt.Sprintf(format, args...)
		b.infoLogger(out)
	}
}
