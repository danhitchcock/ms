package unthermo

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
)

//the filehandle
var filehandle *os.File
//the file itself 
var file io.ReadSeeker
//scanevents contains additional data about the scans (Hz-m/z conversion, scan type, ...)
var scanevents ScanEvents
//scanindexentries is an index containing the scan addresses and additional info
//such as retention time and total current
var scanindexentries ScanIndexEntries
//flag that is true when the whole file resides in memory (faster for Windows machines)
var mem bool

/* Reads the indices from the RAW file in memory
 * it takes multiple options as arguments
 * first option is the file name
 * if the second option is "mem", the raw file will completely be loaded in memory
 */
func Open(options ...string) {
	//Set options
	fn := options[0]
	if len(options) > 1 && options[1] == "mem" {
		mem = true
	}
	
	//Optionally
	if mem {
		//load in memory,
		filecontents, err := ioutil.ReadFile(fn)
		if err != nil {
			log.Fatal("error opening file", err)
		}
		file = bytes.NewReader(filecontents)
	} else {
		//or open file
		filehandle, err := os.Open(fn)
		if err != nil {
			log.Fatal("error opening file", err)
		}
		file = filehandle
	}
	
	//Read headers for file version and RunHeader addresses.
	info, ver := ReadHeaders()
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
	scanindexentries = make(ScanIndexEntries, nScans)
	ReadBetween(rh.ScanindexAddr, rh.ScantrailerAddr, ver, scanindexentries)
	
	//make the offsets absolute in the file instead of relative to the data address
	for i := range scanindexentries {
		scanindexentries[i].Offset += rh.DataAddr
	}
}


/* 
 * Experimental: read out chromatography data from a connected instrument
 */
func Chromatography(instr int) CDataPackets {
	info, ver := ReadHeaders()

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


func Close() error {
	return filehandle.Close()
}

//Read only the initial header part of the file (for the juicy addresses)
func ReadHeaders() (RawFileInfo, Version) {
	hdr := new(FileHeader)
	info := new(RawFileInfo)

	//save position in file after reading, we need to sequentially
	//read some things in order to get to actual byte addresses
	pos := ReadAt(0, 0, hdr)
	ver := hdr.Version

	pos = ReadAt(pos, ver, new(SequencerRow))
	pos = ReadAt(pos, 0, new(AutoSamplerInfo))
	ReadAt(pos, ver, info)

	return *info, ver
}

//Opens a Version v Thermo File, starting at position pos, reads
//data, and returns the position in the file afterwards
func ReadAt(pos uint64, v Version, data Reader) uint64 {
	spos, err := file.Seek(int64(pos), 0)
	if err != nil {
		log.Fatal("error seeking file", err)
	}

	data.Read(file, v)

	spos, err = file.Seek(0, 1)

	if err != nil {
		log.Fatal("error determining position in file", err)
	}
	return uint64(spos)
}

//Reads the range in memory and then fills the Reader
func ReadBetween(begin uint64, end uint64, v Version, data Reader) {
	_, err := file.Seek(int64(begin), 0)
	if err != nil {
		log.Fatal("error seeking file", err)
	}

	b := make([]byte, end-begin) //may fail because of memory requirements
	io.ReadFull(file, b)
	buf := bytes.NewReader(b)

	data.Read(buf, v)
}

//Convert Hz values to m/z
func (data ScanEvent) Convert(v float64) float64 {
	switch data.Nparam {
	case 4:
		return data.A + data.B/v + data.C/v/v
	case 5, 7:
		return data.A + data.B/v/v + data.C/v/v/v/v
	default:
		return v
	}
}
