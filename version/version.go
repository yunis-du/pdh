package version

import (
	"fmt"
	"io"
	"os"
)

var Package = "[pdh]"

var Version = "0.0.1"

func FprintVersion(w io.Writer) {
	_, _ = fmt.Fprintln(w, os.Args[0], Package, Version)
}

func PrintVersion() {
	FprintVersion(os.Stdout)
}
