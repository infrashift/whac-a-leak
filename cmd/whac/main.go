// Command whac is the Whac-a-Leak secret scanner.
package main

import (
	"os"

	"github.com/infrashift/whac-a-leak/internal/cli"
)

func main() {
	os.Exit(cli.Execute(os.Args[1:], os.Stdout, os.Stderr))
}
