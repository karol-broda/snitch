package errutil

import (
	"io"
	"os"

	"github.com/fatih/color"
)

func Ignore[T any](val T, _ error) T {
	return val
}

func IgnoreErr(_ error) {}

func Close(c io.Closer) {
	if c != nil {
		_ = c.Close()
	}
}

// color.Color wrappers - these discard the (int, error) return values

func Print(c *color.Color, a ...any) {
	_, _ = c.Print(a...)
}

func Println(c *color.Color, a ...any) {
	_, _ = c.Println(a...)
}

func Printf(c *color.Color, format string, a ...any) {
	_, _ = c.Printf(format, a...)
}

func Fprintf(c *color.Color, w io.Writer, format string, a ...any) {
	_, _ = c.Fprintf(w, format, a...)
}

// os function wrappers for test cleanup where errors are non-critical

func Setenv(key, value string) {
	_ = os.Setenv(key, value)
}

func Unsetenv(key string) {
	_ = os.Unsetenv(key)
}

func Remove(name string) {
	_ = os.Remove(name)
}

func RemoveAll(path string) {
	_ = os.RemoveAll(path)
}

// Flush calls Flush on a tabwriter and discards the error
type Flusher interface {
	Flush() error
}

func Flush(f Flusher) {
	_ = f.Flush()
}
