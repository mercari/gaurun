package gaurun

import (
	"fmt"
	"runtime"
)

func PrintVersion() {
	fmt.Printf(`Gaurun %s
Compiler: %s %s
Copyright (C) 2014-2019 Mercari, Inc.
`,
		Version,
		runtime.Compiler,
		runtime.Version())

}

func serverHeader() string {
	return fmt.Sprintf("Gaurun %s", Version)
}
