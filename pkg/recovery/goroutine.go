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

var errorHandler func(error) = func(err error) {
	log.Printf("err: %v", err)
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
	r any
}

func newPanicError(r any) PanicError {
	return PanicError{r: r}
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
	} else {
		return "panic: " + fmt.Sprintf("%v", p.r)
	}
}
