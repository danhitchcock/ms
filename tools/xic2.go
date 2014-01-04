/*XIC prints mass chromatograms for given m/z's.

  For every m/z given, it prints the peak with highest intensity in interval
  [mz-tol ppm,mz+tol ppm] for every profile-mode scan.

  Every line contains the mass, retention time and intensity of a peak

  Reading can be sped up on systems with large memory by loading all scans
  in RAM. Add flag -m on the command line for this.

  Example:
      xic -m -mz 361.1466 -mz 445.1200 -tol 2.5 rawfile.raw

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
)

//argument parsing
type mzarg []float64 //a new type for passing flags on the command line

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

func main() {
	var mz mzarg
	var tol float64
	var mem bool
	flag.Var(&mz, "mz", "m/z to filter on, may be specified multiple times")
	flag.Float64Var(&tol, "tol", 0, "allowed m/z tolerance in ppm, can be used with -mz")
	flag.BoolVar(&mem, "m", false, "read all scans in memory for a speed gain")
	flag.Parse()

	xic := func(scan []unthermo.MS) {
		for _, mz := range mz {
			var m []unthermo.MS

			//convert Hz values into m/z and save the signals within range
			for _, ms := range scan {
				if ms.Mz <= mz+10e-6*tol*mz && ms.Mz >= mz-10e-6*tol*mz {
						m = append(m, ms)
					}
			}

			//print the maximum signal of what is saved
			if len(m) > 0 {
				var maxIn unthermo.MS
				for i := range m {
					if m[i].I > maxIn.I {
						maxIn = m[i]
					}
				}
				fmt.Println(mz, maxIn.Time, maxIn.I)
			}
	}	
	}

	for _, filename := range flag.Args() {
		unthermo.OnPeaks(filename, mem, xic)
	}
}
