//XIC prints a "chromatogram" for one ion.
//
//It prints the peak with highest intensity in interval [mz-tol,mz+tol]
//for every profile-mode scan.
//
//Every line contains the retention time and intensity of a peak
//
//Example:
//		xic -mz 361.1466 -tol 0.0025 rawfile.raw
//
//Output:
//		0.003496666666666667 10500.583
//		0.015028333333333333 11793.04
//		0.03391333333333333 10178.598
//		0.05393333333333334 10671.821
//		0.07350833333333334 11572.251
package main

import (
	"bitbucket.org/proteinspector/unthermo"
	"flag"
	"fmt"
)

type MS struct {
	mz float64
	I  float32
}

func main() {
	var mz float64
	var tol float64
	flag.Float64Var(&mz, "mz", 0, "m/z to filter on")
	flag.Float64Var(&tol, "tol", 0, "allowed mz tolerance, can be used with -mz")
	flag.Parse()

	for _, filename := range flag.Args() {
		PrintXIC(filename, mz, tol)
	}
}

func PrintXIC(fn string, mz float64, tol float64) {
	info, ver := unthermo.ReadFileHeaders(fn)

	rh := new(unthermo.RunHeader)
	unthermo.ReadFile(fn, info.Preamble.RunHeaderAddr[0], ver, rh)

	//For later conversion of frequency values to m/z, we need a ScanEvent
	//for each Scan.
	//The list of them starts an uint32 later than ScantrailerAddr
	//the uint32 contains the number of ScanEvents, but we know this
	//already through SampleInfo
	pos := rh.ScantrailerAddr + 4

	//the ScanEvents are of variable size and have no pointer to
	//them, we need to read them all sequentially
	nScans := uint64(rh.SampleInfo.LastScanNumber - rh.SampleInfo.FirstScanNumber + 1)
	scanevents := make([]unthermo.ScanEvent, nScans)
	for i := range scanevents {
		pos = unthermo.ReadFile(fn, pos, ver, &scanevents[i])
	}

	//read all scanindexentries at once, this is probably the fastest
	scanindexentries := make([]unthermo.ScanIndexEntry, nScans)
	for i := range scanindexentries {
		unthermo.ReadFile(fn,
			rh.ScanindexAddr+uint64(i)*scanindexentries[i].Size(ver),
			ver, &scanindexentries[i])
	}

	for s := range scanindexentries {
		scan := new(unthermo.ScanDataPacket)
		unthermo.ReadFile(fn, rh.DataAddr+scanindexentries[s].Offset, 0, scan)

		var m []MS

		//convert the Hz values into m/z and save the signals within range
		for i := uint32(0); i < scan.Profile.PeakCount; i++ {
			for j := uint32(0); j < scan.Profile.Chunks[i].Nbins; j++ {
				tmpmz := scanevents[s].Convert(scan.Profile.FirstValue+
					float64(scan.Profile.Chunks[i].Firstbin+j)*scan.Profile.Step) +
					float64(scan.Profile.Chunks[i].Fudge)
				if tmpmz <= mz+tol && tmpmz >= mz-tol {
					m = append(m, MS{tmpmz,
						scan.Profile.Chunks[i].Signal[j]})
				}
			}
		}

		//print the maximum signal
		if len(m) > 0 {
			var maxInt float32
			for i := range m {
				if m[i].I > maxInt {
					maxInt = m[i].I
				}
			}
			fmt.Println(scanindexentries[s].ScanTime, maxInt)
		}
	}
}
