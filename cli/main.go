package main

import (
	"flag"

	"github.com/oskanberg/gombie"
)

func main() {
	testFlag := flag.String("run", "", "The tests to run. Use the same format as you would with 'go test -run'")
	flag.Parse()

	gombie.Go(*testFlag, gombie.BasicMutators{})
}
