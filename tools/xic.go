/*XIC prints mass chromatograms for given m/z's.

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
	"bitbucket.org/proteinspector/unthermo"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"sort"
)

/*
 * Argument parsing (multiple m/z require a lot of boilerplate)
 */
type mzarg []float64

var mzs mzarg
var tol float64

func (i *mzarg) String() string {
	return fmt.Sprintf("%d", *i)
}

func (i *mzarg) Set(value string) error {
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
 * Actual execution
 */
func main() {
	for _, filename := range flag.Args() {
		//xic gets called on each MS1 Scan read by the unthermo library
		unthermo.Open(filename)
		unthermo.AllScans(xic)
		unthermo.Close()
	}
}

//Algorithm calculating and outputting the peaks belonging to the XIC's
var xic = func(scan unthermo.Scan) {
	//if an MS1 Scan:
	if scan.MSLevel == 1 {
		//for every mz in the argument list
		for _, mz := range mzs {
			//print the peaks within tolerance.
			//The spectrum is sorted by m/z so we can search for the two 
			//border peaks and get the range between them
			lowi := sort.Search(len(scan.Spectrum), func(i int) bool { return scan.Spectrum[i].Mz >= mz-10e-6*tol*mz})
			highi := sort.Search(len(scan.Spectrum), func(i int) bool { return scan.Spectrum[i].Mz >= mz+10e-6*tol*mz})
			//if there is any data in this interval
			if highi > lowi {
				var maxIn unthermo.Peak
				for _, peak := range scan.Spectrum[lowi:highi] {
					if peak.I >= maxIn.I {
						maxIn = peak
					}
				}
				fmt.Println(mz, maxIn.Time, maxIn.I)
			}
		}
	}
}
