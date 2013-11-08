package unthermo

import (
	"encoding/binary"
	"io"
	"log"
	"os"
)


type TrailerLength uint32;
func (data* TrailerLength) Read(r io.Reader, v version) {
	Read(r, data)
}

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
	Read(io.Reader, version)
}

func ReadFile(fn string, pos int64, v version, data Reader) int64 {
	file, err := os.Open(fn)
	pos, err = file.Seek(pos, 0)
	if err != nil {
		log.Fatal(err)
	}

	data.Read(file, v)

	pos, err = file.Seek(0, 1)
	if err != nil {
		log.Fatal(err)
	}
	return pos
}

func (data *ScanEventPreamble) Read(r io.Reader, v version) { 
//128 bytes in v63 and up, 120 in v62, 80 in v57, 41 below that
	switch {
		case v<57:
			*data = make([]uint8, 41)
		case v>=57 && v<62:
			*data = make([]uint8, 80)
		case v>=62 && v<63:
			*data = make([]uint8, 120)
		case v>=63:
			*data = make([]uint8, 128)
	}
	Read(r, data)
}

func (data *ScanEvent) Read(r io.Reader, v version) {
	data.Preamble.Read(r,v)
	Read(r, &data.Nprecursors)
	
	data.Reaction = make([]Reaction, data.Nprecursors)
	for i := range data.Reaction {
		Read(r, &data.Reaction[i])
	}
	
	Read(r, &data.Unknown1)
	Read(r, &data.MZrange)
	Read(r, &data.Nparam)
	
	switch data.Nparam {
		case 4:
			Read(r, &data.Unknown2)
			Read(r, &data.A)
			Read(r, &data.B)
			Read(r, &data.C)
		case 7:
			Read(r, &data.Unknown2)
			Read(r, &data.I)
			Read(r, &data.A)
			Read(r, &data.B)
			Read(r, &data.C)
			Read(r, &data.D)
			Read(r, &data.E)
	}
	
	Read(r, &data.Unknown3)
	Read(r, &data.Unknown4)
}

func (data *FileHeader) Read(r io.Reader, v version) {
	Read(r, data)
}

func (data *PacketHeader) Read(r io.Reader, v version) {
	Read(r, data)
}

func (data *ScanIndexEntry) Read(r io.Reader, v version) {
	if v >= 64 {
		Read(r, data)
	} else {
		Read(r, &data.Offset32)
		Read(r, &data.Index)
		Read(r, &data.Scanevent)
		Read(r, &data.Scansegment)
		Read(r, &data.Next)
		Read(r, &data.Unknown1)
		Read(r, &data.Datasize)
		Read(r, &data.Starttime)
		Read(r, &data.Totalcurrent)
		Read(r, &data.Baseintensity)
		Read(r, &data.Basemz)
		Read(r, &data.Lowmz)
		Read(r, &data.Highmz)
	}
}

func (data *RunHeader) Read(r io.Reader, v version) {
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
	Read(r, &data.Scantrailer_addr32)
	Read(r, &data.Scanparams_addr32)
	Read(r, &data.Unknown3)
	Read(r, &data.Unknown4)
	Read(r, &data.Nsegs)
	Read(r, &data.Unknown5)
	Read(r, &data.Unknown6)
	Read(r, &data.Own_addr32)
	Read(r, &data.Unknown7)
	Read(r, &data.Unknown8)

	if v >= 64 {
		Read(r, &data.Scanindex_addr)
		Read(r, &data.Data_addr)
		Read(r, &data.Instlog_addr)
		Read(r, &data.Errorlog_addr)
		Read(r, &data.Unknown9)
		Read(r, &data.Scantrailer_addr)
		Read(r, &data.Scanparams_addr)
		Read(r, &data.Unknown10)
		Read(r, &data.Own_addr)

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
	}
}

func (data *Info) Read(r io.Reader, v version) {
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
		Read(r, &data.Preamble.Data_addr32)
		Read(r, &data.Preamble.Unknown2)
		Read(r, &data.Preamble.Unknown3)
		Read(r, &data.Preamble.Unknown4)
		Read(r, &data.Preamble.Unknown5)
		Read(r, &data.Preamble.Runheader_addr32)
		if v <= 63 {
			data.Preamble.Unknown6 = make([]byte, 756)
		} else {
			data.Preamble.Unknown6 = make([]byte, 760)
		}
		Read(r, &data.Preamble.Unknown6)
	}
	if v >= 64 {
		Read(r, &data.Preamble.Data_addr)
		Read(r, &data.Preamble.Unknown7)
		Read(r, &data.Preamble.Unknown8)
		Read(r, &data.Preamble.Runheader_addr)

		if v <= 66 {
			data.Preamble.Unknown9 = make([]byte, 1008)
		} else {
			data.Preamble.Unknown9 = make([]byte, 1024)
		}
		Read(r, &data.Preamble.Unknown9)
	}

	Read(r, &data.Heading1)
	Read(r, &data.Heading2)
	Read(r, &data.Heading3)
	Read(r, &data.Heading4)
	Read(r, &data.Heading5)
	Read(r, &data.Unknown1)
}

func (data *AutoSamplerInfo) Read(r io.Reader, v version) {
	Read(r, &data.Preamble)
	Read(r, &data.Text)
}

func (data *SequencerRow) Read(r io.Reader, v version) {
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
