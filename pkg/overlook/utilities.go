package overlook

import "io"

// CheckClose used with defer
func CheckClose(v interface{}) {
	if d, ok := v.(io.Closer); ok {
		err := d.Close()
		if err != nil {
			panic(err)
		}
	}
}
