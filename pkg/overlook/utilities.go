package overlook

import (
	"io"
	"os"
)

// CheckClose used with defer
func CheckClose(v interface{}) {
	if d, ok := v.(io.Closer); ok {
		err := d.Close()
		if err != nil {
			panic(err)
		}
	}
}

// Exists reports whether the named file or directory exists.
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
