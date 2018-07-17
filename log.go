package xmlrpc

// LogErrorFunc can be set from outside the library to allow error logging
var LogErrorFunc func(string, ...interface{})

func logError(format string, args ...interface{}) {
	if LogErrorFunc == nil {
		return
	}

	LogErrorFunc(format, args...)
}
