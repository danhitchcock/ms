package main

import (
	"bitbucket.org/proteinspector/unthermo"
	"fmt"
	"os"
	"strconv"
)

func main() {
	//Parse arguments
	filename := os.Args[1]
	//Execute Scan with argument prettyprint
	info, ver := ReadFileHeaders(filename)
}
