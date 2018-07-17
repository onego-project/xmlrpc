package xmlrpc

import "github.com/pkg/errors"

type temporary interface {
	Temporary() bool // Is the error temporary?
}

type timeout interface {
	Timeout() bool // Is the error a timeout?
}

// IsTemporary returns true if err is temporary.
func IsTemporary(err error) bool {
	te, ok := errors.Cause(err).(temporary)
	return ok && te.Temporary()
}

// IsTimeout returns true if err is caused by timeout.
func IsTimeout(err error) bool {
	te, ok := errors.Cause(err).(timeout)
	return ok && te.Timeout()
}
