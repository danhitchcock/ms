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
	var mem bool
	flag.Float64Var(&mz, "mz", 0, "m/z to filter on")
	flag.Float64Var(&tol, "tol", 0, "allowed mz tolerance, can be used with -mz")
	flag.BoolVar(&mem, "m", false, "read all scans in memory for a speed gain")
	flag.Parse()

	for _, filename := range flag.Args() {
		PrintXIC(filename, mz, tol, mem)
	}
}

func PrintXIC(fn string, mz float64, tol float64, mem bool) {
	info, ver := unthermo.ReadFileHeaders(fn)

	rh := new(unthermo.RunHeader)
	unthermo.ReadFile(fn, info.Preamble.RunHeaderAddr[0], ver, rh)

	//For later conversion of frequency values to m/z, we need a ScanEvent
	//for each Scan.
	//The list of them starts an uint32 later than ScantrailerAddr
	nScans := uint64(rh.SampleInfo.LastScanNumber - rh.SampleInfo.FirstScanNumber + 1)
	scanevents := make(unthermo.Scanevents, nScans)
	unthermo.ReadFileRange(fn, rh.ScantrailerAddr + 4, rh.ScanparamsAddr, ver, scanevents)
	
	//read all scanindexentries (for retention time) at once,
	//this is probably the fastest
	scanindexentries := make(unthermo.ScanIndexEntries, nScans)
	unthermo.ReadFileRange(fn, rh.ScanindexAddr, rh.ScantrailerAddr, ver, scanindexentries)

	//initialize the shared data structure
	
	if mem {
		//create channel to share memory with library
		ch := make(chan *unthermo.ScanDataPacket)
		//send off library to wait for work
		go unthermo.ReadScansFromMemory(fn, rh.DataAddr, rh.OwnAddr, 0, ch)
	
		for s:=uint64(0); s<nScans ; s++ {
			scan := new(unthermo.ScanDataPacket)
			ch <- scan 	//send pointer to data structure
			scan = <-ch //receive pointer back when library is done
			PrintMaxInt(scan, &scanevents[s], &scanindexentries[s], mz, tol)
		}
	} else {
		for s := range scanindexentries {
			scan := new(unthermo.ScanDataPacket)
			unthermo.ReadFile(fn, rh.DataAddr+scanindexentries[s].Offset, 0, scan)
			PrintMaxInt(scan, &scanevents[s], &scanindexentries[s], mz, tol)
		}
	}
}

func PrintMaxInt(scan *unthermo.ScanDataPacket, scanevent *unthermo.ScanEvent, sie *unthermo.ScanIndexEntry, mz float64, tol float64) {
	var m []MS

	//convert Hz values into m/z and save the signals within range
	for i := uint32(0); i < scan.Profile.PeakCount; i++ {
		for j := uint32(0); j < scan.Profile.Chunks[i].Nbins; j++ {
			tmpmz := scanevent.Convert(scan.Profile.FirstValue+
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
		var maxIn MS
		for i := range m {
			if m[i].I > maxIn.I {
				maxIn = m[i]
			}
		}
		fmt.Println(sie.Index, maxIn.mz, maxIn.I)
	}
}
