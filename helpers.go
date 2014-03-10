package unthermo

import (
	"bytes"
	"io"
	"log"
	"os"
)

//interface shared by all data objects in the raw file
type Reader interface {
	Read(io.Reader, Version)
}

//Opens a Version v Thermo File, starting at position pos, reads
//data, and returns the position in the file afterwards
func ReadFile(fn string, pos uint64, v Version, data Reader) uint64 {
	file, err := os.Open(fn)
	if err != nil {
		log.Fatal("error opening file", err)
	}
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
