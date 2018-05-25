package hpfeeds

type Logger func(...interface{})

var debugLog Logger = nil

func SetDebugLogger(logger Logger) {
	debugLog = logger
}

func logDebug(args ...interface{}) {
	if debugLog != nil {
		debugLog(args)
	}
}

var errorLog Logger = nil

func SetErrorLogger(logger Logger) {
	errorLog = logger
}

func logError(args ...interface{}) {
	if errorLog != nil {
		errorLog(args)
	}
}
