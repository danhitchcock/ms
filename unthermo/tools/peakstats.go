/*The XIC tool prints mass chromatograms for a specific m/z.

  For the m/z given, it prints the peak with highest intensity in interval
  [mz-tol ppm,mz+tol ppm] for every MS-1 scan.

  Every line contains the retention time and intensity of a peak

  Example:
      xic -mz 361.1466 -tol 2.5 -raw rawfile.raw

  Output:
      361.1466 0.003496666666666667 10500.583
      361.1466 0.015028333333333333 11793.04
      361.1466 0.03391333333333333 10178.598
      361.1466 0.05393333333333334 10671.821
      361.1466 0.07350833333333334 11572.251
*/
package main

import (
	"bitbucket.org/proteinspector/ms"
	"bitbucket.org/proteinspector/ms/unthermo"
	"flag"
	"fmt"
	"log"
	"github.com/pkelchte/spline"
	"sort"
)


//ions are the m/z for the XIC
//var ions = []float64{495.78700, 424.25560, 507.81340, 461.74760, 740.40170, 820.47250, 682.34770} //BSA
var ions = []float64{363.67450, 362.22910, 367.21590, 550.76660, 643.85824, 878.47842, 789.90439} //Enolase
//tol is the tolerance in ppm
var tol float64 = 2.5

type TimedPeak struct {
	ms.Peak
	Time float64 
}

/*
  The peakstats tool outputs a few data about the peaks of supplied ions:
  - Mass, Time and Intensity of maximum peak in LC/MS map
  - Full Width Half Max of this peak
  - Optionally: an interpolation of the peak data itself (for graphing)
*/
func main() {
	var fileName string
	flag.StringVar(&fileName, "raw", "small.RAW", "name of the subject RAW file")
	flag.Parse()
	
	file, err := unthermo.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	xicmap := xics(file)
	resolution := guessMsOneInterval(file)
	
	for k := range xicmap {
		fmt.Println(k, fwhm(xicmap[k], resolution))
	}
}

//guessMsOneInterval returns a guess for the interval between ms1 scans
func guessMsOneInterval(file unthermo.File) float64 {
	var timeOne float64 = 0
	var i int = 1
	for ; timeOne == 0; i++ {
		scan := file.Scan(i)
		if scan.MSLevel == 1 {
			timeOne = scan.Time
		}
	}
	timeTwo := timeOne
	for ; timeTwo == timeOne; i++ {
		scan := file.Scan(i)
		if scan.MSLevel == 1 {
			timeTwo = scan.Time
		}
	}
	return timeTwo - timeOne
}

func fwhm(xic []TimedPeak, resolution float64) float64 {
	
	s := spline.Spline{}

	X := make([]float64, len(xic))
	Y := make([]float64, len(xic))
	for i:= range xic {
		X[i] = xic[i].Time
		Y[i] = float64(xic[i].I)
	}

	s.Set_points(X, Y, true)
	
	var max TimedPeak
	
	for _, peak := range xic {
		if peak.I >= max.I {
			max = peak
		}
	}
	
	right := max.Time
	for ; s.Operate(right)>float64(max.I/2); right += resolution {} //increase with seconds
	left := max.Time
	for ; s.Operate(left)>float64(max.I/2); left-=resolution {}
	
	//for i := -13; i< 44; i++ {
		//x := 0.01 * float64(i)
		//fmt.Printf("%f %f\n", x, s.Operate(x))
	//}
	
	return right - left
}

//xics returns a map of slices of extracted ion chromatgrams
func xics(file unthermo.File) map[float64][]TimedPeak {
	sort.Float64s(ions)
	
	xicmap := make(map[float64][]TimedPeak, len(ions))
		
	for i := 1; i <= file.NScans(); i++ {
		scan := file.Scan(i)
	
		if scan.MSLevel == 1 {
			spectrum := scan.Spectrum()
			
			for _, ion := range ions {
				filteredSpectrum := mzFilter(spectrum, ion, tol)
				if len(filteredSpectrum) > 0 {
					xicmap[ion] = append(xicmap[ion],TimedPeak{maxPeak(filteredSpectrum), scan.Time})
				}
			}
		}
	}
	return xicmap
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
