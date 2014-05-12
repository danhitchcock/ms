/*The XIC tool prints mass chromatograms for specific m/z's.

  For every m/z given, it prints the peak with highest intensity in interval
  [mz-tol ppm,mz+tol ppm] for every MS-1 scan.

  Every line contains the mass, retention time and intensity of a peak

  Example:
      xic -mz 361.1466 -mz 445.1200 -tol 2.5 rawfile.raw

  Output:
      361.1466 0.003496666666666667 10500.583
      445.12 0.003496666666666667 37872.473
      361.1466 0.015028333333333333 11793.04
      445.12 0.015028333333333333 41592.734
      361.1466 0.03391333333333333 10178.598
      445.12 0.03391333333333333 38692.445
      361.1466 0.05393333333333334 10671.821
      445.12 0.05393333333333334 37769.496
      361.1466 0.07350833333333334 11572.251
      445.12 0.07350833333333334 37978.258
*/
package main

import (
	"bitbucket.org/proteinspector/ms"
	"bitbucket.org/proteinspector/ms/unthermo"
	"flag"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
)

/*
 * Argument parsing (multiple m/z require a lot of boilerplate)
 */
type floatList []float64

//mzs is the list of mz's for the XIC
var mzs floatList

//tol is the tolerance in ppm
var tol float64

func (i *floatList) String() string {
	return fmt.Sprintf("%d", *i)
}

func (i *floatList) Set(value string) error {
	for _, splv := range strings.Split(value, ",") {
		tmp, err := strconv.ParseFloat(splv, 64)
		if err != nil {
			return err
		}
		*i = append(*i, tmp)
	}
	return nil
}

func init() {
	flag.Var(&mzs, "mz", "m/z to filter on, this flag may be specified multiple times")
	flag.Float64Var(&tol, "tol", 0, "allowed m/z tolerance in ppm, can be used with -mz")
	flag.Parse()
}

/*
  Actual execution, where XIC peaks get extracted out of each MS1 Scan
  read by the unthermo library
*/
func main() {
	for _, filename := range flag.Args() {
		file, err := unthermo.Open(filename)
		if err != nil {
			log.Println(err)
		}

		for i := 1; i <= file.NScans(); i++ {
			XICpeaks(file.Scan(i), mzs, tol)
		}

		file.Close()
	}
}

//XIC outputs the scan time and peaks of a MS1 scan within tolerance
//around the supplied mzs
func XICpeaks(scan ms.Scan, mzs []float64, tol float64) {
	//if an MS1 Scan:
	if scan.MSLevel == 1 {
		spectrum := scan.Spectrum()
		//for every mz in the argument list
		for _, mz := range mzs {
			//filter around the mz
			filteredSpectrum := mzFilter(spectrum, mz, tol)
			//If there is any data in this interval
			if len(filteredSpectrum) > 0 {
				//look for the maximal Peak
				maxIn := maxPeak(filteredSpectrum)
				//print it.
				fmt.Println(mz, scan.Time, maxIn.I)
			}
		}
	}
}

//maxPeak returns the maximally intense peak within the supplied spectrum
func maxPeak(spectrum ms.Spectrum) (maxIn ms.Peak) {
	//find the peak with maximal intensity
	for _, peak := range spectrum {
		if peak.I >= maxIn.I {
			maxIn = peak
		}
	}
	return
}

//mzFilter outputs the spectrum within tolerance around the supplied mz
func mzFilter(spectrum ms.Spectrum, mz float64, tol float64) ms.Spectrum {
	//A spectrum is sorted by m/z so we can search for the two
	//border peaks and get the range between them.
	lowi := sort.Search(len(spectrum), func(i int) bool { return spectrum[i].Mz >= mz-10e-6*tol*mz })
	highi := sort.Search(len(spectrum), func(i int) bool { return spectrum[i].Mz >= mz+10e-6*tol*mz })
	//If there is any data in this interval
	return spectrum[lowi:highi]
}
