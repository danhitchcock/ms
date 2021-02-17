//The printspectrum tool prints out the spectrum (mz and intensity values) of a
//Thermo RAW File
//
//  Every line of the output is a peak registered by the mass spectrometer
//  characterized by an m/z value in Da and an intensity in the mass spectrometer's unit of abundance
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/danhitchcock/ms"
	"github.com/danhitchcock/ms/unthermo"
)

func main() {
	var scannumber int
	var filename string

	//Parse arguments
	flag.IntVar(&scannumber, "sn", 1, "the scan number")
	flag.StringVar(&filename, "raw", "small.RAW", "name of the RAW file")
	flag.Parse()

	//open RAW file
	file, err := unthermo.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	//Print the Spectrum at the supplied scan number
	printspectrum(file.Scan(scannumber))
}

//Print m/z and Intensity of every peak in the spectrum
func printspectrum(scan ms.Scan) {
	for _, peak := range scan.Spectrum() {
		fmt.Println(peak.Mz, peak.I)
	}
}
