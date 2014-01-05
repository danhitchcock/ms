package unthermo

type Peak struct {
	Time float64
	Mz   float64
	I    float32
}

type Spectrum []Peak

//On every encountered MS Scan, the function fun is called
func OnAllScans(fn string, mem bool, mslev uint8, fun func(Spectrum)) {
	//Read necessary headers
	info, ver := ReadFileHeaders(fn)
	rh := new(RunHeader)
	ReadFile(fn, info.Preamble.RunHeaderAddr[0], ver, rh)

	//For later conversion of frequency values to m/z, we need a ScanEvent
	//for each Scan.
	//The list of them starts an uint32 later than ScantrailerAddr
	nScans := uint64(rh.SampleInfo.LastScanNumber - rh.SampleInfo.FirstScanNumber + 1)
	sevs := make(Scanevents, nScans)
	ReadFileRange(fn, rh.ScantrailerAddr+4, rh.ScanparamsAddr, ver, sevs)

	//read all scanindexentries (for retention time) at once,
	//this is probably the fastest
	sies := make(ScanIndexEntries, nScans)
	ReadFileRange(fn, rh.ScanindexAddr, rh.ScantrailerAddr, ver, sies)

	if mem {
		//create channels to share memory with library
		offset := make(chan uint64)
		scans := make(chan *ScanDataPacket)

		//send off library to wait for work
		go ReadScansFromMemory(fn, rh.DataAddr, rh.OwnAddr, 0, offset, scans)

		for i, sie := range sies {
			if sevs[i].Preamble[6] == mslev {
				offset <- sie.Offset //send location of data structure
				scan := <-scans      //receive pointer back when library is done
				OnScan(scan, &sevs[i], &sie, fun)
			}
		}
	} else {
		for i, sie := range sies {
			if sevs[i].Preamble[6] == mslev {
				scan := new(ScanDataPacket)
				ReadFile(fn, rh.DataAddr+sie.Offset, 0, scan)
				OnScan(scan, &sevs[i], &sie, fun)
			}
		}
	}
}

func OnScan(rawscan *ScanDataPacket, scanevent *ScanEvent,
	sie *ScanIndexEntry, fun func(Spectrum)) {

	var spectrum Spectrum

	//convert Hz values into m/z and save the profile peaks
	for i := uint32(0); i < rawscan.Profile.PeakCount; i++ {
		for j := uint32(0); j < rawscan.Profile.Chunks[i].Nbins; j++ {
			tmpmz := scanevent.Convert(rawscan.Profile.FirstValue+
				float64(rawscan.Profile.Chunks[i].Firstbin+j)*rawscan.Profile.Step) +
				float64(rawscan.Profile.Chunks[i].Fudge)
			spectrum = append(spectrum,
				Peak{Time: sie.Time, Mz: tmpmz, I: rawscan.Profile.Chunks[i].Signal[j]})
		}
	}

	//Also save the Centroided Peaks (they also occur in profile scans!?)
	for i := uint32(0); i < rawscan.PeakList.Count; i++ {
		spectrum = append(spectrum,
			Peak{Time: sie.Time, Mz: float64(rawscan.PeakList.Peaks[i].Mz),
				I: rawscan.PeakList.Peaks[i].Abundance})
	}

	fun(spectrum)
}
