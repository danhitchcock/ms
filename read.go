package unthermo

import (
	"encoding/binary"
	"io"
	"log"
	"os"
	"math"
)

func Read(r io.Reader, data interface{}) {
	switch v := data.(type) {
	case *PascalString:
		binary.Read(r, binary.LittleEndian, &v.Length)
		v.Text = make([]uint16, v.Length)
		binary.Read(r, binary.LittleEndian, &v.Text)
	default:
		binary.Read(r, binary.LittleEndian, v)
	}
}

type Reader interface {
	Read(io.Reader, Version)
}

//Reads a Version v Thermo File, starting at position pos, returns the
//position in the file after the read
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

func ReadHeaders(fn string) (RawFileInfo, Version) {
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

type TrailerLength uint32

func (data *TrailerLength) Read(r io.Reader, v Version) {
	Read(r, data)
}

func (data *ScanDataPacket) Read(r io.Reader, v Version) {
	Read(r, &data.Header)

	if data.Header.ProfileSize > 0 {
		Read(r, &data.Profile.FirstValue)
		Read(r, &data.Profile.Step)
		Read(r, &data.Profile.PeakCount)
		Read(r, &data.Profile.Nbins)

		data.Profile.Chunks = make([]ProfileChunk, data.Profile.PeakCount)
		for i := range data.Profile.Chunks {
			Read(r, &data.Profile.Chunks[i].Firstbin)
			Read(r, &data.Profile.Chunks[i].Nbins)
			if data.Header.Layout > 0 {
				Read(r, &data.Profile.Chunks[i].Fudge)
			}
			data.Profile.Chunks[i].Signal = make([]float32, data.Profile.Chunks[i].Nbins)
			
			for j := range data.Profile.Chunks[i].Signal {
				buf := make([]byte, 4)
				if _, err := io.ReadFull(r, buf); err != nil {
					log.Fatal("error reading scan: ", err)
				}
				data.Profile.Chunks[i].Signal[j]= math.Float32frombits(uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24)
			}
		}
	}

	if data.Header.PeaklistSize > 0 {
		Read(r, &data.PeakList.Count)
		data.PeakList.Peaks = make([]Peak, data.PeakList.Count)
		for i := range data.PeakList.Peaks {
			Read(r, &data.PeakList.Peaks[i])
		}
	}
}

func (data *ScanEvent) Read(r io.Reader, v Version) {
	if v<66 {
		switch {
	case v < 57:
		Read(r, data.Preamble[:41])
	case v >= 57 && v < 62:
		Read(r, data.Preamble[:80])
	case v >= 62 && v < 63:
		Read(r, data.Preamble[:120])
	case v >= 63:
		Read(r, data.Preamble[:128])
	}
	Read(r, &data.Nprecursors)
	data.Reaction = make([]Reaction, data.Nprecursors)
	for i := range data.Reaction {
		Read(r, &data.Reaction[i])
	}

	Read(r, &data.Unknown1[0])
	Read(r, &data.MZrange[0])
	Read(r, &data.Nparam)

	switch data.Nparam {
	case 4:
		Read(r, &data.Unknown2[0])
		Read(r, &data.A)
		Read(r, &data.B)
		Read(r, &data.C)
	case 7:
		Read(r, data.Unknown2[0:2])
		Read(r, &data.A)
		Read(r, &data.B)
		Read(r, &data.C)
		Read(r, data.Unknown2[2:4])
	}

	Read(r, data.Unknown1[1:3])
	} else { //v66
		Read(r, &data.Preamble)
		Read(r, &data.Unknown1[0])
		Read(r, &data.Nprecursors) //this is just a guess according to Gene Selkov
		if data.Preamble[10] == 1 { //ms2 (dependent scan)
			data.Reaction = make([]Reaction, data.Nprecursors)
			for i := range data.Reaction {
				Read(r, &data.Reaction[i])
			}
			Read(r, data.Unknown2[0:2])
			Read(r, data.Unknown1[1:4])
			Read(r, &data.MZrange[0])
			Read(r, &data.Nparam)
		} else { //ms1
			Read(r, &data.MZrange[0])
			Read(r, data.Unknown1[1:5])
			Read(r, &data.MZrange[1])
			Read(r, data.Unknown1[5:8])
			Read(r, &data.MZrange[2])
			Read(r, &data.Nparam)
		}
		Read(r, data.Unknown2[2:4])
		Read(r, &data.A)
		Read(r, &data.B)
		Read(r, &data.C)
		Read(r, data.Unknown1[8:13])
	}
}

func (data *FileHeader) Read(r io.Reader, v Version) {
	Read(r, data)
}

func (data *CDataPacket) Read(r io.Reader, v Version) {
	Read(r, data)
}

func (data CIndexEntry) Size(v Version) uint64 {
	switch {
	case v < 64:
		return 64
	default:
		return 72
	}
}

func (data *CIndexEntry) Read(r io.Reader, v Version) {
	switch {
	case v < 64:
		Read(r, &data.Offset32)
		Read(r, &data.Index)
		Read(r, &data.Event)
		Read(r, &data.Unknown1)
		Read(r, &data.Unknown2)
		Read(r, &data.Unknown3)
		Read(r, &data.Unknown4)
		Read(r, &data.Unknown5)
		Read(r, &data.Time)
		Read(r, &data.Unknown6)
		Read(r, &data.Unknown7)
		Read(r, &data.Value)

		data.Offset = uint64(data.Offset32)
	default:
		Read(r, data)
	}

}

func (data ScanIndexEntry) Size(v Version) uint64 {
	switch {
	case v < 64:
		return 72
	case v == 64:
		return 80
	default:
		return 88
	}
}

func (data *ScanIndexEntry) Read(r io.Reader, v Version) {
	if v == 66 {
		Read(r, data)
	} else if v == 64 {
		Read(r, &data.Offset32)
		Read(r, &data.Index)
		Read(r, &data.Scanevent)
		Read(r, &data.Scansegment)
		Read(r, &data.Next)
		Read(r, &data.Unknown1)
		Read(r, &data.DataPacketSize)
		Read(r, &data.ScanTime)
		Read(r, &data.Totalcurrent)
		Read(r, &data.Baseintensity)
		Read(r, &data.Basemz)
		Read(r, &data.Lowmz)
		Read(r, &data.Highmz)
		Read(r, &data.Offset)
	} else {
		Read(r, &data.Offset32)
		Read(r, &data.Index)
		Read(r, &data.Scanevent)
		Read(r, &data.Scansegment)
		Read(r, &data.Next)
		Read(r, &data.Unknown1)
		Read(r, &data.DataPacketSize)
		Read(r, &data.ScanTime)
		Read(r, &data.Totalcurrent)
		Read(r, &data.Baseintensity)
		Read(r, &data.Basemz)
		Read(r, &data.Lowmz)
		Read(r, &data.Highmz)

		data.Offset = uint64(data.Offset32)
	}
}

func (data *RunHeader) Read(r io.Reader, v Version) {
	Read(r, &data.SampleInfo)
	Read(r, &data.Filename1)
	Read(r, &data.Filename2)
	Read(r, &data.Filename3)
	Read(r, &data.Filename4)
	Read(r, &data.Filename5)
	Read(r, &data.Filename6)
	Read(r, &data.Unknown1)
	Read(r, &data.Unknown2)
	Read(r, &data.Filename7)
	Read(r, &data.Filename8)
	Read(r, &data.Filename9)
	Read(r, &data.Filename10)
	Read(r, &data.Filename11)
	Read(r, &data.Filename12)
	Read(r, &data.Filename13)
	Read(r, &data.ScantrailerAddr32)
	Read(r, &data.ScanparamsAddr32)
	Read(r, &data.Unknown3)
	Read(r, &data.Unknown4)
	Read(r, &data.Nsegs)
	Read(r, &data.Unknown5)
	Read(r, &data.Unknown6)
	Read(r, &data.OwnAddr32)
	Read(r, &data.Unknown7)
	Read(r, &data.Unknown8)

	data.ScanindexAddr = uint64(data.SampleInfo.ScanindexAddr)
	data.DataAddr = uint64(data.SampleInfo.DataAddr)
	data.InstlogAddr = uint64(data.SampleInfo.InstlogAddr)
	data.ErrorlogAddr = uint64(data.SampleInfo.ErrorlogAddr)
	data.ScantrailerAddr = uint64(data.ScantrailerAddr32)
	data.ScanparamsAddr = uint64(data.ScanparamsAddr32)

	if v >= 64 {
		Read(r, &data.ScanindexAddr)
		Read(r, &data.DataAddr)
		Read(r, &data.InstlogAddr)
		Read(r, &data.ErrorlogAddr)
		Read(r, &data.Unknown9)
		Read(r, &data.ScantrailerAddr)
		Read(r, &data.ScanparamsAddr)
		Read(r, &data.Unknown10)
		Read(r, &data.Unknown11)
		Read(r, &data.OwnAddr)

		Read(r, &data.Unknown12)
		Read(r, &data.Unknown13)
		Read(r, &data.Unknown14)
		Read(r, &data.Unknown15)
		Read(r, &data.Unknown16)
		Read(r, &data.Unknown17)
		Read(r, &data.Unknown18)
		Read(r, &data.Unknown19)
		Read(r, &data.Unknown20)
		Read(r, &data.Unknown21)
		Read(r, &data.Unknown22)
		Read(r, &data.Unknown23)
		Read(r, &data.Unknown24)
		Read(r, &data.Unknown25)
		Read(r, &data.Unknown26)
		Read(r, &data.Unknown27)
		Read(r, &data.Unknown28)
		Read(r, &data.Unknown29)
		Read(r, &data.Unknown30)
		Read(r, &data.Unknown31)
		Read(r, &data.Unknown32)
		Read(r, &data.Unknown33)
		Read(r, &data.Unknown34)
		Read(r, &data.Unknown35)
	}

	Read(r, &data.Unknown36)
	Read(r, &data.Unknown37)
	Read(r, &data.Device)
	Read(r, &data.Model)
	Read(r, &data.SN)
	Read(r, &data.SWVer)
	Read(r, &data.Tag1)
	Read(r, &data.Tag2)
	Read(r, &data.Tag3)
	Read(r, &data.Tag4)
}

func (data *RawFileInfo) Read(r io.Reader, v Version) {
	Read(r, &data.Preamble.Methodfilepresent)
	Read(r, &data.Preamble.Year)
	Read(r, &data.Preamble.Month)
	Read(r, &data.Preamble.Weekday)
	Read(r, &data.Preamble.Day)
	Read(r, &data.Preamble.Hour)
	Read(r, &data.Preamble.Minute)
	Read(r, &data.Preamble.Second)
	Read(r, &data.Preamble.Millisecond)

	if v >= 57 {
		Read(r, &data.Preamble.Unknown1)
		Read(r, &data.Preamble.DataAddr32)
		Read(r, &data.Preamble.NControllers)
		Read(r, &data.Preamble.NControllers2)
		Read(r, &data.Preamble.Unknown2)
		Read(r, &data.Preamble.Unknown3)
		if v < 64 {
			data.Preamble.RunHeaderAddr32 = make([]uint32, data.Preamble.NControllers)
			data.Preamble.Unknown4 = make([]uint32, data.Preamble.NControllers)
			data.Preamble.Unknown5 = make([]uint32, data.Preamble.NControllers)
			for i := range data.Preamble.RunHeaderAddr32 {
				Read(r, &data.Preamble.RunHeaderAddr32[i])
				Read(r, &data.Preamble.Unknown4[i])
				Read(r, &data.Preamble.Unknown5[i])
			}

			data.Preamble.RunHeaderAddr = make([]uint64, data.Preamble.NControllers)
			for i := range data.Preamble.RunHeaderAddr {
				data.Preamble.RunHeaderAddr[i] = uint64(data.Preamble.RunHeaderAddr32[i])
			}

			if v == 57 {
				Read(r, data.Preamble.Padding1[:756-12*data.Preamble.NControllers])
			} else {
				Read(r, data.Preamble.Padding1[:760-12*data.Preamble.NControllers])
			}
		} else {
			Read(r, &data.Preamble.Padding1)
		}

	}
	if v >= 64 {
		Read(r, &data.Preamble.DataAddr)
		Read(r, &data.Preamble.Unknown6)

		data.Preamble.RunHeaderAddr = make([]uint64, data.Preamble.NControllers)
		data.Preamble.Unknown7 = make([]uint64, data.Preamble.NControllers)
		for i := range data.Preamble.RunHeaderAddr {
			Read(r, &data.Preamble.RunHeaderAddr[i])
			Read(r, &data.Preamble.Unknown7[i])
		}
		if v < 66 {
			Read(r, data.Preamble.Padding2[:1016-16*data.Preamble.NControllers])
		} else {
			Read(r, data.Preamble.Padding2[:1032-16*data.Preamble.NControllers])
		}
	}

	Read(r, &data.Heading1)
	Read(r, &data.Heading2)
	Read(r, &data.Heading3)
	Read(r, &data.Heading4)
	Read(r, &data.Heading5)
	Read(r, &data.Unknown1)
}

func (data *AutoSamplerInfo) Read(r io.Reader, v Version) {
	Read(r, &data.Preamble)
	Read(r, &data.Text)
}

func (data *SequencerRow) Read(r io.Reader, v Version) {
	Read(r, &data.Injection)

	Read(r, &data.Unknown1)
	Read(r, &data.Unknown2)
	Read(r, &data.Id)
	Read(r, &data.Comment)
	Read(r, &data.Userlabel1)
	Read(r, &data.Userlabel2)
	Read(r, &data.Userlabel3)
	Read(r, &data.Userlabel4)
	Read(r, &data.Userlabel5)
	Read(r, &data.Instmethod)
	Read(r, &data.Procmethod)
	Read(r, &data.Filename)
	Read(r, &data.Path)

	if v >= 57 {
		Read(r, &data.Vial)
		Read(r, &data.Unknown3)
		Read(r, &data.Unknown4)
		Read(r, &data.Unknown5)
	}
	if v >= 60 {
		Read(r, &data.Unknown6)
		Read(r, &data.Unknown7)
		Read(r, &data.Unknown8)
		Read(r, &data.Unknown9)
		Read(r, &data.Unknown10)
		Read(r, &data.Unknown11)
		Read(r, &data.Unknown12)
		Read(r, &data.Unknown13)
		Read(r, &data.Unknown14)
		Read(r, &data.Unknown15)
		Read(r, &data.Unknown16)
		Read(r, &data.Unknown17)
		Read(r, &data.Unknown18)
		Read(r, &data.Unknown19)
		Read(r, &data.Unknown20)
	}
}
