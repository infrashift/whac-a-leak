package cli

import (
	"fmt"
	"io"
)

// printer writes formatted output and keeps the first write error, so a
// command can print many lines and check for failure once.
type printer struct {
	w   io.Writer
	err error
}

func (p *printer) printf(format string, a ...any) {
	if p.err == nil {
		_, p.err = fmt.Fprintf(p.w, format, a...)
	}
}
