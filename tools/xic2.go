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

type MS struct {
	mz float64
	I  float32
}

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

	xic := func(scan *unthermo.ScanDataPacket,
		scanevent *unthermo.ScanEvent, sie *unthermo.ScanIndexEntry) {
		for _, mz := range mz {
			var m []MS

			//convert Hz values into m/z and save the signals within range
			for i := uint32(0); i < scan.Profile.PeakCount; i++ {
				for j := uint32(0); j < scan.Profile.Chunks[i].Nbins; j++ {
					tmpmz := scanevent.Convert(scan.Profile.FirstValue+
						float64(scan.Profile.Chunks[i].Firstbin+j)*scan.Profile.Step) +
						float64(scan.Profile.Chunks[i].Fudge)
					if tmpmz <= mz+10e-6*tol*mz && tmpmz >= mz-10e-6*tol*mz {
						m = append(m, MS{tmpmz,
							scan.Profile.Chunks[i].Signal[j]})
					}
				}
			}

			//print the maximum signal
			if len(m) > 0 {
				var maxIn MS
				for i := range m {
					if m[i].I > maxIn.I {
						maxIn = m[i]
					}
				}
				fmt.Println(mz, sie.Time, maxIn.I)
			}
		}
	}

	for _, filename := range flag.Args() {
		XIC(filename, mz, tol, mem)
	}
}

func XIC(fn string, mz mzarg, tol float64, mem bool) {
	//Read necessary headers
	info, ver := unthermo.ReadFileHeaders(fn)
	rh := new(unthermo.RunHeader)
	unthermo.ReadFile(fn, info.Preamble.RunHeaderAddr[0], ver, rh)

	//For later conversion of frequency values to m/z, we need a ScanEvent
	//for each Scan.
	//The list of them starts an uint32 later than ScantrailerAddr
	nScans := uint64(rh.SampleInfo.LastScanNumber - rh.SampleInfo.FirstScanNumber + 1)
	scanevents := make(unthermo.Scanevents, nScans)
	unthermo.ReadFileRange(fn, rh.ScantrailerAddr+4, rh.ScanparamsAddr, ver, scanevents)

	//read all scanindexentries (for retention time) at once,
	//this is probably the fastest
	scanindexentries := make(unthermo.ScanIndexEntries, nScans)
	unthermo.ReadFileRange(fn, rh.ScanindexAddr, rh.ScantrailerAddr, ver, scanindexentries)

	if mem {
		//create channels to share memory with library
		offset := make(chan uint64)
		scans := make(chan *unthermo.ScanDataPacket)

		//send off library to wait for work
		go unthermo.ReadScansFromMemory(fn, rh.DataAddr, rh.OwnAddr, 0, offset, scans)

		for i, s := range scanindexentries {
			offset <- s.Offset //send location of data structure
			scan := <-scans    //receive pointer back when library is done
			for _, mz := range mz {
				PrintMaxPeak(scan, &scanevents[i], &s, mz, tol)
			}
		}
	} else {
		for i := range scanindexentries {
			scan := new(unthermo.ScanDataPacket)
			unthermo.ReadFile(fn, rh.DataAddr+scanindexentries[i].Offset, 0, scan)
			for _, mz := range mz {
				PrintMaxPeak(scan, &scanevents[i], &scanindexentries[i], mz, tol)
			}
		}
	}
}

func PrintMaxPeak(scan *unthermo.ScanDataPacket, scanevent *unthermo.ScanEvent, sie *unthermo.ScanIndexEntry, mz float64, tol float64) {
	var m []MS

	//convert Hz values into m/z and save the signals within range
	for i := uint32(0); i < scan.Profile.PeakCount; i++ {
		for j := uint32(0); j < scan.Profile.Chunks[i].Nbins; j++ {
			tmpmz := scanevent.Convert(scan.Profile.FirstValue+
				float64(scan.Profile.Chunks[i].Firstbin+j)*scan.Profile.Step) +
				float64(scan.Profile.Chunks[i].Fudge)
			if tmpmz <= mz+10e-6*tol*mz && tmpmz >= mz-10e-6*tol*mz {
				m = append(m, MS{tmpmz,
					scan.Profile.Chunks[i].Signal[j]})
			}
		}
	}

	//print the maximum signal
	if len(m) > 0 {
		var maxIn MS
		for i := range m {
			if m[i].I > maxIn.I {
				maxIn = m[i]
			}
		}
		fmt.Println(mz, sie.Time, maxIn.I)
	}
}
