package unthermo


type Peak struct {
	Time float64
	Mz   float64
	I    float32
}

type Spectrum []Peak

type Scan struct {
	Spectrum
	MSLevel uint8
	Activation Activation
}
	
type Activation int

const (
        CID Activation = iota
        HCD
)

//Spectrum implements sort.Interface for []Peak based on m/z
func (a Spectrum) Len() int           { return len(a) }
func (a Spectrum) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Spectrum) Less(i, j int) bool { return a[i].Mz < a[j].Mz }

/*
 * Convenience function that runs over all spectra in the raw file
 * 
 * On every encountered MS Scan, the function fun is called
 */
func AllScans(fun func(Scan)) {	
	for i, sie := range scanindexentries {
			scn := new(ScanDataPacket)
			ReadAt(sie.Offset, 0, scn)
			fun(scan(scn, &scanevents[i], &sie))
		}
}

func ScanAt(sn uint64) Scan {
	//read Scan Packet for the above scan number
	scn := new(ScanDataPacket)
	ReadAt(scanindexentries[sn-1].Offset, 0, scn)

	return scan(scn, &scanevents[sn-1], &scanindexentries[sn-1])
}

func scan(rawscan *ScanDataPacket, scanevent *ScanEvent,
	sie *ScanIndexEntry) Scan {

	var scan Scan

	scan.MSLevel = scanevent.Preamble[6]

	if rawscan.Profile.PeakCount > 0 {
		//convert Hz values into m/z and save the profile peaks
		for i := uint32(0); i < rawscan.Profile.PeakCount; i++ {
			for j := uint32(0); j < rawscan.Profile.Chunks[i].Nbins; j++ {
				tmpmz := scanevent.Convert(rawscan.Profile.FirstValue+
					float64(rawscan.Profile.Chunks[i].Firstbin+j)*rawscan.Profile.Step) +
					float64(rawscan.Profile.Chunks[i].Fudge)
				scan.Spectrum = append(scan.Spectrum,
					Peak{Time: sie.Time, Mz: tmpmz, I: rawscan.Profile.Chunks[i].Signal[j]})
			}
		}
	} else {
		//Save the Centroided Peaks, they also occur in profile scans but
		//overlap with profiles, Thermo always does centroiding just for fun
		for i := uint32(0); i < rawscan.PeakList.Count; i++ {
			scan.Spectrum = append(scan.Spectrum,
				Peak{Time: sie.Time, Mz: float64(rawscan.PeakList.Peaks[i].Mz),
					I: rawscan.PeakList.Peaks[i].Abundance})
		}
	}

	return scan
}


