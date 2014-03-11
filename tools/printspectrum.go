package main

import (
	"bitbucket.org/proteinspector/unthermo"
	"fmt"
	"os"
	"strconv"
)

func main() {
	//Parse arguments
	scannumber, _ := strconv.Atoi(os.Args[1])
	filename := os.Args[2]
	unthermo.Open(filename)
	//Execute Scan with argument prettyprint
	prettyprint(unthermo.ScanAt(uint64(scannumber)))
}

//Print mz and Intensity of every peak in spectrum
func prettyprint(scan unthermo.Scan) {
	for _, peak := range scan.Spectrum {
		fmt.Println(peak.Mz, peak.I)
	}
}
