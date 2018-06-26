package hpfeeds

import (
	"fmt"
)

type Logger func(...interface{})

func (b *Broker) SetDebugLogger(logger Logger) {
	b.debugLog = logger
}

func (b *Broker) logDebug(args ...interface{}) {
	if b.debugLog != nil {
		b.debugLog(args)
	}
}

func (b *Broker) logDebugf(format string, args ...interface{}) {
	if b.infoLog != nil {
		out := fmt.Sprintf(format, args)
		b.infoLog(out)
	}
}

func (b *Broker) SetErrorLogger(logger Logger) {
	b.errorLog = logger
}

func (b *Broker) logError(args ...interface{}) {
	if b.errorLog != nil {
		b.errorLog(args)
	}
}

func (b *Broker) logErrorf(format string, args ...interface{}) {
	if b.infoLog != nil {
		out := fmt.Sprintf(format, args)
		b.infoLog(out)
	}
}

func (b *Broker) SetInfoLogger(logger Logger) {
	b.infoLog = logger
}

func (b *Broker) logInfo(args ...interface{}) {
	if b.infoLog != nil {
		b.infoLog(args)
	}
}

func (b *Broker) logInfof(format string, args ...interface{}) {
	if b.infoLog != nil {
		out := fmt.Sprintf(format, args)
		b.infoLog(out)
	}
}
