package gaurun

import (
	"fmt"
	"runtime"
)

func PrintGaurunVersion() {
	fmt.Printf(`Gaurun %s
Compiler: %s %s
Copyright (C) 2014-2015 Mercari, Inc.
`,
		Version,
		runtime.Compiler,
		runtime.Version())

}

func serverHeader() string {
	return fmt.Sprintf("Gaurun %s", Version)
}
