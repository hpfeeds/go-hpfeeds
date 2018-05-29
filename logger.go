package hpfeeds

type Logger func(...interface{})

func (b *Broker) SetDebugLogger(logger Logger) {
	b.debugLog = logger
}

func (b *Broker) logDebug(args ...interface{}) {
	if b.debugLog != nil {
		debugLog(args)
	}
}

func (b *Broker) SetErrorLogger(logger Logger) {
	b.errorLog = logger
}

func (b *Broker) logError(args ...interface{}) {
	if b.errorLog != nil {
		errorLog(args)
	}
}
