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
	//open RAW file
	unthermo.Open(filename)
	//Print the Spectrum at the supplied scan number
	printspectrum(unthermo.ScanNumber(scannumber))
	unthermo.Close()
}

//Print mz and Intensity of every peak in the spectrum
func printspectrum(scan unthermo.Scan) {
	for _, peak := range scan.Spectrum {
		fmt.Println(peak.Mz, peak.I)
	}
}
