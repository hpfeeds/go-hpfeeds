package hpfeeds

import (
	"fmt"
)

type Logger func(...interface{})

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
