//The printspectrum tool prints out the spectrum (mz and intensity values) of a
//Thermo RAW File
//
//  Every line of the output is a peak registered by the mass spectrometer
//  characterized by an m/z value in Da and an intensity in the mass spectrometer's unit of abundance
package main

import (
	"bitbucket.org/proteinspector/ms"
	"bitbucket.org/proteinspector/ms/unthermo"
	"fmt"
	"os"
	"strconv"
)

func main() {
	//Parse arguments
	scannumber, _ := strconv.Atoi(os.Args[1])
	filename := os.Args[2]

	//open RAW file
	rf, _ := unthermo.Open(filename)

	//Print the Spectrum at the supplied scan number
	printspectrum(rf.Scan(scannumber))

	rf.Close()
}

//Print m/z and Intensity of every peak in the spectrum
func printspectrum(scan ms.Scan) {
	for _, peak := range scan.Spectrum {
		fmt.Println(peak.Mz, peak.I)
	}
}
