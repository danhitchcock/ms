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
	"log"
	"flag"
	"sort"
)

var reporter_ions = [...]float64 { 316.1200, 581.883 }
const tol = 2.5 //ppm
var emptyPeak = ms.Peak{0, 0}
var filename string

func main() {
	//Parse arguments
	flag.StringVar(&filename, "raw", "small.RAW", "name of the RAW file")
	flag.Parse()

	//open RAW file
	file, err := unthermo.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	printExtendedCIDScans(file)

}


//@pre The MS2 scans have one precursor
func printExtendedCIDScans(file unthermo.File) {
	cidSpectra := make(map[float64]ms.Spectrum)
	hcdPeakSpectra := make(map[float64]ms.Spectrum)
	var msTwoHappened bool = false
	
	for i := 1; i <= file.NScans(); i++ {
		scan := file.Scan(i)
		switch scan.MSLevel {
			case 1:
				if msTwoHappened {
					if len(cidSpectra) != len(hcdPeakSpectra) {
						log.Println("unmatched spectra right before scan", i)
						break
					}
					
					for precursor := range hcdPeakSpectra {
						mergeAndPrint(cidSpectra[precursor], hcdPeakSpectra[precursor])
					}
					
					cidSpectra = make(map[float64]ms.Spectrum)
					hcdPeakSpectra = make(map[float64]ms.Spectrum)
					msTwoHappened = false
				}
			case 2:
				msTwoHappened = true
				switch scan.Analyzer {
					case ms.FTMS:
						spectrum := scan.Spectrum()
						var hcdPeaks ms.Spectrum
						
						for _, mz := range reporter_ions {
							peak := peak(spectrum, mz, tol)
							if peak != emptyPeak {
								hcdPeaks = append(hcdPeaks, peak)
							}
						}
						
						hcdPeakSpectra[scan.PrecursorMzs[0]] = hcdPeaks
					case ms.ITMS:
						cidSpectra[scan.PrecursorMzs[0]] = scan.Spectrum()
				}
		}
	}
}

func mergeAndPrint(left ms.Spectrum, right ms.Spectrum) {
	left = append(left, right...)
	sort.Sort(left)
    printMGFSpectrum(left)
}

//Print m/z and Intensity of every peak in the spectrum
func printMGFSpectrum(spectrum ms.Spectrum) {
	for _, peak := range spectrum {
		fmt.Println(peak.Mz, peak.I)
	}
}

//the highest MS1 peak with mz within tolerance around the supplied mz.
func peak(spectrum ms.Spectrum, mz float64, tol float64) (peak ms.Peak) {
		filteredSpectrum := mzFilter(spectrum, mz, tol)
		if len(filteredSpectrum) > 0 {
			return maxPeak(filteredSpectrum)
		} else {
			return peak
		}
}

//mzFilter outputs the spectrum within tol ppm around the supplied mz
func mzFilter(spectrum ms.Spectrum, mz float64, tol float64) ms.Spectrum {
	return mzIntervalFilter(spectrum, mz-10e-6*tol*mz, mz+10e-6*tol*mz)
}

//mzIntervalFilter filters the spectrum for mz's within the interval [min,max)
//including minMz and excluding maxMZ
func mzIntervalFilter(spectrum ms.Spectrum, minMz float64, maxMz float64) ms.Spectrum {
	//A spectrum is sorted by m/z so we can do binary search for two
	//endpoint peaks and get the peaks between them.
	lowi := sort.Search(len(spectrum), func(i int) bool { return spectrum[i].Mz >= minMz })
	highi := sort.Search(len(spectrum)-lowi, func(i int) bool { return spectrum[i+lowi].Mz >= maxMz })

	return spectrum[lowi : highi+lowi]
}

//maxPeak returns the maximally intense peak within the supplied spectrum
func maxPeak(spectrum ms.Spectrum) (maxIn ms.Peak) {
	for _, peak := range spectrum {
		if peak.I >= maxIn.I {
			maxIn = peak
		}
	}
	return
}
