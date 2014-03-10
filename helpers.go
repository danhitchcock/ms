package unthermo

import (
	"bytes"
	"io"
	"log"
	"os"
)

//the file itself 
var file io.Reader
//scanevents is the general index for the scans
var scanevents ScanEvents
//scanindexentries is an additional index containing more info about the scans
var scanindexentries ScanIndexEntries
//flag that is true when the whole file resides in memory (faster for Windows machines)
var mem bool

/* Reads the indices from the RAW file in memory
 * it takes multiple options as arguments
 * first option is the file name
 * if the second option is "mem", the raw file will completely be loaded in memory
 */
func Open(options ...string) {
	//Set options:
	fn := options[0]
	if options[1] == "mem" {
		mem = true
	}
	
	
	//Read headers for file version and RunHeader addresses.
	info, ver := ReadFileHeaders(fn)
	rh := new(RunHeader)

	//read runheaders until we have a non-empty Scantrailer Address
	//indicating it is the runheader for a MS device (not a chromatography device)
	for i := 0; i < len(info.Preamble.RunHeaderAddr) && rh.ScantrailerAddr == 0; i++ {
		ReadFile(fn, info.Preamble.RunHeaderAddr[i], ver, rh)
	}
	if rh.ScantrailerAddr == 0 {
		log.Fatal("Couldn't find MS run header in file at positions ", info.Preamble.RunHeaderAddr)
	}

	//For later conversion of frequency values to m/z, we need a ScanEvent
	//for each Scan.
	//The list of them starts an uint32 later than ScantrailerAddr
	nScans := uint64(rh.SampleInfo.LastScanNumber - rh.SampleInfo.FirstScanNumber + 1)
	scanevents = make(ScanEvents, nScans)
	ReadFileRange(fn, rh.ScantrailerAddr+4, rh.ScanparamsAddr, ver, scanevents)

	//read all scanindexentries (for retention time) at once,
	scanindexentries = make(ScanIndexEntries, nScans)
	ReadFileRange(fn, rh.ScanindexAddr, rh.ScantrailerAddr, ver, scanindexentries)
}

//Opens a Version v Thermo File, starting at position pos, reads
//data, and returns the position in the file afterwards
func ReadFile(pos uint64, v Version, data Reader) uint64 {
	
	defer file.Close()
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

//Read only the initial header part of the file (for the juicy addresses)
func ReadFileHeaders(fn string) (RawFileInfo, Version) {
	hdr := new(FileHeader)
	info := new(RawFileInfo)

	//save position in file after reading, we need to sequentially
	//read some things in order to get to actual byte addresses
	pos := ReadFile(fn, 0, 0, hdr)
	ver := hdr.Version

	pos = ReadFile(fn, pos, ver, new(SequencerRow))
	pos = ReadFile(fn, pos, 0, new(AutoSamplerInfo))
	ReadFile(fn, pos, ver, info)

	return *info, ver
}

//Reads the range in memory and then fills the Reader, for faster reads
func ReadFileRange(fn string, begin uint64, end uint64, v Version, data Reader) {
	file, err := os.Open(fn)
	if err != nil {
		log.Fatal("error opening file", err)
	}
	defer file.Close()
	_, err = file.Seek(int64(begin), 0)
	if err != nil {
		log.Fatal("error seeking file", err)
	}

	b := make([]byte, end-begin) //may fail because of memory requirements
	io.ReadFull(file, b)
	buf := bytes.NewReader(b)

	data.Read(buf, v)
}

//Reads the range in memory, then waits for scan packet offsets
//when the scan packet is read, send it back on the out channel
func ReadScansFromMemory(fn string, begin uint64, end uint64, v Version, in <-chan uint64, out chan<- *ScanDataPacket) {
	file, err := os.Open(fn)
	if err != nil {
		log.Fatal("error opening file", err)
	}
	defer file.Close()
	_, err = file.Seek(int64(begin), 0)
	if err != nil {
		log.Fatal("error seeking file", err)
	}

	b := make([]byte, end-begin) //may fail because of memory requirements
	io.ReadFull(file, b)
	buf := bytes.NewReader(b)

	//for each incoming scan packet offset
	for offset := range in {
		scan := new(ScanDataPacket)
		buf.Seek(int64(offset), 0)
		scan.Read(buf, v) //read the file at that offset
		out <- scan       //and send a reference to the corresponding scan back
	}
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
