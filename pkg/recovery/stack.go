package recovery

import (
	"runtime"
	"runtime/debug"
)

const maxStretch = 3

// debug.Stack() dynamically grows the buffer size without any limits.
// To prevent uncontrolled buffer growth, this implementation limits the growth to maxStretch (3 times).
// The maximum buffer size is capped at 8192KB.
func GetStackTrace() string {
	buf := make([]byte, 1024)
	stretchCnt := 1
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return string(buf[:n])
		}

		if stretchCnt > maxStretch {
			return string(buf)
		}

		buf = make([]byte, 2*len(buf))
		stretchCnt++

		debug.Stack()
	}
}
