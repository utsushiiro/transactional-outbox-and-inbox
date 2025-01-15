package recovery

import (
	"fmt"
	"log"
	"sync"
)

func Go(fn func() error) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicErr := newPanicError(r)
				errorHandler(panicErr)
			}
		}()

		err := fn()
		if err != nil {
			errorHandler(err)
		}
	}()
}

var errorHandler = func(err error) {
	log.Printf("err: %+v", err)
}
var errorHandlerOpsLock = sync.Mutex{}

func SetErrorHandler(fn func(error)) {
	errorHandlerOpsLock.Lock()
	defer errorHandlerOpsLock.Unlock()
	if fn != nil {
		errorHandler = fn
	}
}

type PanicError struct {
	r          any
	stackTrace string
}

func newPanicError(r any) PanicError {
	return PanicError{
		r:          r,
		stackTrace: GetStackTrace(),
	}
}

func (p PanicError) Unwrap() error {
	if err, ok := p.r.(error); ok {
		return err
	}

	return nil
}

func (p PanicError) Error() string {
	if err := p.Unwrap(); err != nil {
		return "panic: " + err.Error()
	}

	return "panic: " + fmt.Sprintf("%v", p.r)
}

func (p PanicError) StackTrace() string {
	return p.stackTrace
}

func (p PanicError) GoString() string {
	return fmt.Sprintf("PanicError{r: %#v, stackTrace: %q}", p.r, p.stackTrace)
}

func (p PanicError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			// Include the stack trace for %+v
			fmt.Fprintf(s, "%s\n\n%s", p.Error(), p.stackTrace)
		case s.Flag('#'):
			// Go-syntax representation(GoString)
			fmt.Fprint(s, p.GoString())
		default:
			fmt.Fprint(s, p.Error())
		}
	case 's':
		// Default string representation
		fmt.Fprint(s, p.Error())
	case 'q':
		// Quoted string representation
		fmt.Fprintf(s, "%q", p.Error())
	}
}
