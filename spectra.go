package unthermo

type MS struct {
	Time float64
	Mz float64
	I  float32
}

func OnPeaks(fn string, mem bool, fun func([]MS)) {
	//Read necessary headers
	info, ver := ReadFileHeaders(fn)
	rh := new(RunHeader)
	ReadFile(fn, info.Preamble.RunHeaderAddr[0], ver, rh)

	//For later conversion of frequency values to m/z, we need a ScanEvent
	//for each Scan.
	//The list of them starts an uint32 later than ScantrailerAddr
	nScans := uint64(rh.SampleInfo.LastScanNumber - rh.SampleInfo.FirstScanNumber + 1)
	scanevents := make(Scanevents, nScans)
	ReadFileRange(fn, rh.ScantrailerAddr+4, rh.ScanparamsAddr, ver, scanevents)

	//read all scanindexentries (for retention time) at once,
	//this is probably the fastest
	scanindexentries := make(ScanIndexEntries, nScans)
	ReadFileRange(fn, rh.ScanindexAddr, rh.ScantrailerAddr, ver, scanindexentries)

	if mem {
		//create channels to share memory with library
		offset := make(chan uint64)
		scans := make(chan *ScanDataPacket)

		//send off library to wait for work
		go ReadScansFromMemory(fn, rh.DataAddr, rh.OwnAddr, 0, offset, scans)

		for i, s := range scanindexentries {
			offset <- s.Offset //send location of data structure
			scan := <-scans    //receive pointer back when library is done
			OnProfile(scan, &scanevents[i], &s, fun)
		}
	} else {
		for i := range scanindexentries {
			scan := new(ScanDataPacket)
			ReadFile(fn, rh.DataAddr+scanindexentries[i].Offset, 0, scan)
			OnProfile(scan, &scanevents[i], &scanindexentries[i], fun)
		}
	}
}


func OnProfile(rawscan *ScanDataPacket, scanevent *ScanEvent,
 sie *ScanIndexEntry, fun func([]MS)) {
		
		var scan []MS
		
			//convert Hz values into m/z and save the signals within range
			for i := uint32(0); i < rawscan.Profile.PeakCount; i++ {
				for j := uint32(0); j < rawscan.Profile.Chunks[i].Nbins; j++ {
					tmpmz := scanevent.Convert(rawscan.Profile.FirstValue+
						float64(rawscan.Profile.Chunks[i].Firstbin+j)*rawscan.Profile.Step) +
						float64(rawscan.Profile.Chunks[i].Fudge)
							scan = append(scan,
							MS{Time: sie.Time, Mz: tmpmz, I:
							rawscan.Profile.Chunks[i].Signal[j]})
					}
				}
			
		fun(scan)
		
	}
