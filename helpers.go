package unthermo

import (
	"log"
	"os"
)

//the Thermo RAW file
var file *os.File
//scanevents contains additional data about the scans (Hz-m/z conversion, scan type, ...)
var scanevents ScanEvents
//scanindexentries is an index containing the scan addresses and additional info
//such as retention time and total current
var scanindex ScanIndex
//flag that is true when the whole file resides in memory (faster for Windows machines)
var mem bool

// Opens the supplied filename and reads the indices from the RAW file in memory
func Open(fn string) {
	var err error
	file, err = os.Open(fn)
	if err != nil {
		log.Fatal("error opening file", err)
	}
	
	//Read headers for file version and RunHeader addresses.
	info, ver := readHeaders()
	rh := new(RunHeader)

	//read runheaders until we have a non-empty Scantrailer Address
	//indicating it is the runheader for a MS device (not a chromatography device)
	for i := 0; i < len(info.Preamble.RunHeaderAddr) && rh.ScantrailerAddr == 0; i++ {
		ReadAt(info.Preamble.RunHeaderAddr[i], ver, rh)
	}
	if rh.ScantrailerAddr == 0 {
		log.Fatal("Couldn't find MS run header in file at positions ", info.Preamble.RunHeaderAddr)
	}

	//For later conversion of frequency values to m/z, we need a ScanEvent
	//for each Scan.
	//The list of them starts an uint32 later than ScantrailerAddr
	nScans := uint64(rh.SampleInfo.LastScanNumber - rh.SampleInfo.FirstScanNumber + 1)
	scanevents = make(ScanEvents, nScans)
	ReadBetween(rh.ScantrailerAddr+4, rh.ScanparamsAddr, ver, scanevents)

	//read all scanindexentries at once
	scanindex = make(ScanIndex, nScans)
	ReadBetween(rh.ScanindexAddr, rh.ScantrailerAddr, ver, scanindex)
	
	//make the offsets absolute in the file instead of relative to the data address
	for i := range scanindex {
		scanindex[i].Offset += rh.DataAddr
	}
}

//Close the RAW file
func Close() error {
	return file.Close()
}

/* 
 * Experimental: read out chromatography data from a connected instrument
 */
func Chromatography(instr int) CDataPackets {
	info, ver := readHeaders()

	if uint32(instr) > info.Preamble.NControllers-1 {
		log.Fatal(instr, " is higher than number of extra controllers: ", info.Preamble.NControllers-1)
	}
	
	rh := new(RunHeader)
	ReadAt(info.Preamble.RunHeaderAddr[instr], ver, rh)
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
		ReadAt(rh.DataAddr+i*16, ver, &cdata[i]) //16 bytes of CDataPacket
	}
	return cdata
}

