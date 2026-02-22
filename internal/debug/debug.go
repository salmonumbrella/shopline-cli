package debug

import (
	"fmt"
	"io"
	"time"
)

// Logger provides debug logging.
type Logger struct {
	w io.Writer
}

// New creates a new debug logger.
func New(w io.Writer) *Logger {
	return &Logger{w: w}
}

// Nop returns a logger that discards all output.
func Nop() *Logger {
	return &Logger{w: io.Discard}
}

// Printf logs a debug message.
func (l *Logger) Printf(format string, args ...interface{}) {
	if l.w == nil {
		return
	}
	timestamp := time.Now().Format("15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	_, _ = fmt.Fprintf(l.w, "[DEBUG] %s %s\n", timestamp, msg)
}
