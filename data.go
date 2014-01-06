package unthermo

import (
	"log"
	"sort"
)

type Peak struct {
	Time float64
	Mz   float64
	I    float32
}

type Spectrum []Peak

//Spectrum implements sort.Interface for []Peak based on m/z
func (a Spectrum) Len() int           { return len(a) }
func (a Spectrum) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Spectrum) Less(i, j int) bool { return a[i].Mz < a[j].Mz }

//On every encountered MS Scan, the function fun is called
func AllScans(fn string, mem bool, mslev uint8, fun func(Spectrum)) {
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
				scn := <-scans       //receive pointer back when library is done
				scan(scn, &sevs[i], &sie, fun)
			}
		}
	} else {
		for i, sie := range sies {
			if sevs[i].Preamble[6] == mslev {
				scn := new(ScanDataPacket)
				ReadFile(fn, rh.DataAddr+sie.Offset, 0, scn)
				scan(scn, &sevs[i], &sie, fun)
			}
		}
	}
}

func scan(rawscan *ScanDataPacket, scanevent *ScanEvent,
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

	sort.Sort(spectrum)
	fun(spectrum)
}

func Scan(fn string, sn uint64, fun func(Spectrum)) {
	info, ver := ReadFileHeaders(fn)

	rh := new(RunHeader)
	ReadFile(fn, info.Preamble.RunHeaderAddr[0], ver, rh)

	//the MS RunHeader contains besides general info three interesting
	//addresses: ScanindexAddr (with the scan headers), DataAddr,
	//and ScantrailerAddr (which includes orbitrap Hz-m/z conversion
	//parameters and info about the scans)

	if sn < uint64(rh.SampleInfo.FirstScanNumber) || sn > uint64(rh.SampleInfo.LastScanNumber) {
		log.Fatal("scan number out of range: ", rh.SampleInfo.FirstScanNumber, ", ", rh.SampleInfo.LastScanNumber)
	}

	//read the n'th ScanIndexEntry
	sie := new(ScanIndexEntry)
	ReadFile(fn, rh.ScanindexAddr+(sn-1)*sie.Size(ver), ver, sie)

	//For later conversion of frequency values to m/z, we need a ScanEvent
	//The list of them starts 4 bytes later than ScantrailerAddr
	pos := rh.ScantrailerAddr + 4

	//the ScanEvents are of variable size and have no pointer to
	//them, we need to read at least all the ones preceding n
	scanevent := new(ScanEvent)
	for i := uint64(0); i < sn; i++ {
		pos = ReadFile(fn, pos, ver, scanevent)
	}

	//read Scan Packet for the above scan number
	scn := new(ScanDataPacket)
	ReadFile(fn, rh.DataAddr+sie.Offset, 0, scn)

	scan(scn, scanevent, sie, fun)
}

//@pre instr>0. in other words: not the mass spectrometer
func Chromatography(fn string, instr int, fun func(CDataPackets)) {
	info, ver := ReadFileHeaders(fn)

	if uint32(instr) > info.Preamble.NControllers-1 {
		log.Fatal(instr, " is higher than number of extra controllers: ", info.Preamble.NControllers-1)
	}

	rh := new(RunHeader)
	ReadFile(fn, info.Preamble.RunHeaderAddr[instr], ver, rh)

	//The instrument RunHeader contains an interesting address: DataAddr
	//There is another address ScanIndexAddr, which points to CIndexEntry
	//containers at ScanIndexAddr. Less data can be read for now

	nScan := uint64(rh.SampleInfo.LastScanNumber - rh.SampleInfo.FirstScanNumber + 1)
	cdata := make(CDataPackets, nScan)
	for i := uint64(0); i < nScan; i++ {
		ReadFile(fn, rh.DataAddr+i*16, ver, &cdata[i]) //16 bytes of CDataPacket
	}

	fun(cdata)
}
