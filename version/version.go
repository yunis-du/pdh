package version

import (
	"fmt"
	"io"
	"os"
)

var Package = "[pdh]"

var Version = "0.1.1"

func FprintVersion(w io.Writer) {
	_, _ = fmt.Fprintln(w, Package, Version)
}

func PrintVersion() {
	FprintVersion(os.Stdout)
}
