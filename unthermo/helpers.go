package unthermo

import (
	"errors"
	"os"
	"bitbucket.org/proteinspector/ms"
)

type File struct {
	//the file on disk
	f *os.File
	//scanevents contains additional data about the scans (Hz-m/z conversion, scan type, ...)
	scanevents ScanEvents
	//scanindexentries is an index containing the scan addresses and additional info
	//such as retention time and total current
	scanindex ScanIndex
}

//Opens the supplied filename and reads the indices from the RAW file in memory. Multiple files may be read concurrently.
func Open(fn string) (file File, err error) {
	f, err := os.Open(fn)
	if err != nil {
		return
	}
	
	//Read headers for file version and RunHeader addresses.
	info, ver := readHeaders(f)
	rh := new(RunHeader)

	//read runheaders until we have a non-empty Scantrailer Address
	//indicating it is the runheader for a MS device (not a chromatography device)
	for i := 0; i < len(info.Preamble.RunHeaderAddr) && rh.ScantrailerAddr == 0; i++ {
		readAt(f, info.Preamble.RunHeaderAddr[i], ver, rh)
	}
	if rh.ScantrailerAddr == 0 {
		err = errors.New("Couldn't find MS run header in file")
		return
	}

	//For later conversion of frequency values to m/z, we need a ScanEvent
	//for each Scan.
	//The list of them starts an uint32 later than ScantrailerAddr
	nScans := uint64(rh.SampleInfo.LastScanNumber - rh.SampleInfo.FirstScanNumber + 1)
	scanevents := make(ScanEvents, nScans)
	readBetween(f, rh.ScantrailerAddr+4, rh.ScanparamsAddr, ver, scanevents)

	//read all scanindexentries at once
	scanindex := make(ScanIndex, nScans)
	readBetween(f, rh.ScanindexAddr, rh.ScantrailerAddr, ver, scanindex)
	
	//make the offsets absolute in the file instead of relative to the data address
	for i := range scanindex {
		scanindex[i].Offset += rh.DataAddr
	}
	
	return File{f: f, scanevents: scanevents, scanindex: scanindex}, err
}

//Close the RAW file
func (rf File) Close() error {
	return rf.f.Close()
}

/* 
 * Experimental: read out chromatography data from a connected instrument
 */
func (rf File) Chromatography(instr int) CDataPackets {
	info, ver := readHeaders(rf.f)

	if uint32(instr) > info.Preamble.NControllers-1 {
		log.Fatal(instr, " is higher than number of extra controllers: ", info.Preamble.NControllers-1)
	}
	
	rh := new(RunHeader)
	readAt(rf.f, info.Preamble.RunHeaderAddr[instr], ver, rh)
	//The ScantrailerAddr has to be 0. in other words: we're not looking at the MS runheader
	if rh.ScantrailerAddr != 0 {
		log.Fatal("You selected the MS instrument, no chromatography data can be read.")
	}
	
	//The instrument RunHeader contains an interesting address: DataAddr
	//There is another address ScanIndexAddr, which points to CIndexEntry
	//containers at ScanIndexAddr. Less data can be read for now

	nScan := uint64(rh.SampleInfo.LastScanNumber - rh.SampleInfo.FirstScanNumber + 1)
	cdata := make(CDataPackets, nScan)
	for i := uint64(0); i < nScan; i++ {
		readAt(rf.f, rh.DataAddr+i*16, ver, &cdata[i]) //16 bytes of CDataPacket
	}
	return cdata
}

/*
 * Convenience function that runs over all spectra in the raw file
 *
 * On every encountered MS Scan, the function fun is called
 */
func (rf *File) AllScans(fun func(ms.Scan)) {
	for i, sie := range rf.scanindex {
		scn := new(ScanDataPacket)
		readBetween(rf.f, sie.Offset, sie.Offset+uint64(sie.DataPacketSize), 0, scn)
		fun(scan(scn, &rf.scanevents[i], &sie))
	}
}

//Returns the number of scans in the index
func (rf *File) NScans() int {
	return len(rf.scanindex)
}

func (rf *File) Scan(sn int) ms.Scan {
	if sn < 1 || sn > rf.NScans() {
		log.Fatal("Scan Number ", sn, " is out of bounds [1, ", rf.NScans(), "]")
	}

	//read Scan Packet for the above scan number
	scn := new(ScanDataPacket)
	readBetween(rf.f, rf.scanindex[sn-1].Offset, rf.scanindex[sn-1].Offset+uint64(rf.scanindex[sn-1].DataPacketSize), 0, scn)
	
	return scan(scn, &rf.scanevents[sn-1], &rf.scanindex[sn-1])
}

/*
 * Converts the three Thermo scan data structures into a general structure
 */
func scan(rawscan *ScanDataPacket, scanevent *ScanEvent,
	sie *ScanIndexEntry) ms.Scan {

	var scan ms.Scan

	scan.MSLevel = scanevent.Preamble[6]

	if rawscan.Profile.PeakCount > 0 {
		//convert Hz values into m/z and save the profile peaks
		for i := uint32(0); i < rawscan.Profile.PeakCount; i++ {
			for j := uint32(0); j < rawscan.Profile.Chunks[i].Nbins; j++ {
				tmpmz := scanevent.Convert(rawscan.Profile.FirstValue+
					float64(rawscan.Profile.Chunks[i].Firstbin+j)*rawscan.Profile.Step) +
					float64(rawscan.Profile.Chunks[i].Fudge)
				scan.Spectrum = append(scan.Spectrum,
					ms.Peak{Time: sie.Time, Mz: tmpmz, I: rawscan.Profile.Chunks[i].Signal[j]})
			}
		}
	} else {
		//Save the Centroided Peaks, they also occur in profile scans but
		//overlap with profiles, Thermo always does centroiding just for fun
		for i := uint32(0); i < rawscan.PeakList.Count; i++ {
			scan.Spectrum = append(scan.Spectrum,
				ms.Peak{Time: sie.Time, Mz: float64(rawscan.PeakList.Peaks[i].Mz),
					I: rawscan.PeakList.Peaks[i].Abundance})
		}
	}

	return scan
}

