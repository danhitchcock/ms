package unthermo

import (
	"encoding/binary"
	"io"
	"log"
	"os"
)

func ReadPacketHeader(fn string, pos int64, v version) (PacketHeader, int64) {
	file, err := os.Open(fn)
	pos, err = file.Seek(pos, 0)
	if err != nil {
		log.Fatal(err)
	}

	data := new(PacketHeader)
	ThermoRead(file, data)

	pos, err = file.Seek(0, 1)
	if err != nil {
		log.Fatal(err)
	}
	return *data, pos
}

func ReadScanIndexEntry(fn string, pos int64, v version) (ScanIndexEntry, int64) {
	file, err := os.Open(fn)
	pos, err = file.Seek(pos, 0)
	if err != nil {
		log.Fatal(err)
	}

	data := new(ScanIndexEntry)
	if v >= 64 {
		ThermoRead(file, data)
	} else {
		ThermoRead(file, &data.Offset32)
		ThermoRead(file, &data.Index)
		ThermoRead(file, &data.Scanevent)
		ThermoRead(file, &data.Scansegment)
		ThermoRead(file, &data.Next)
		ThermoRead(file, &data.Unknown1)
		ThermoRead(file, &data.Datasize)
		ThermoRead(file, &data.Starttime)
		ThermoRead(file, &data.Totalcurrent)
		ThermoRead(file, &data.Baseintensity)
		ThermoRead(file, &data.Basemz)
		ThermoRead(file, &data.Lowmz)
		ThermoRead(file, &data.Highmz)
	}

	pos, err = file.Seek(0, 1)
	if err != nil {
		log.Fatal(err)
	}
	return *data, pos
}

func ReadRunHeader(fn string, pos int64, v version) (RunHeader, int64) {
	file, err := os.Open(fn)
	pos, err = file.Seek(pos, 0)
	if err != nil {
		log.Fatal(err)
	}

	data := new(RunHeader)

	ThermoRead(file, &data.SampleInfo)
	ThermoRead(file, &data.Filename1)
	ThermoRead(file, &data.Filename2)
	ThermoRead(file, &data.Filename3)
	ThermoRead(file, &data.Filename4)
	ThermoRead(file, &data.Filename5)
	ThermoRead(file, &data.Filename6)
	ThermoRead(file, &data.Unknown1)
	ThermoRead(file, &data.Unknown2)
	ThermoRead(file, &data.Filename7)
	ThermoRead(file, &data.Filename8)
	ThermoRead(file, &data.Filename9)
	ThermoRead(file, &data.Filename10)
	ThermoRead(file, &data.Filename11)
	ThermoRead(file, &data.Filename12)
	ThermoRead(file, &data.Filename13)
	ThermoRead(file, &data.Scantrailer_addr32)
	ThermoRead(file, &data.Scanparams_addr32)
	ThermoRead(file, &data.Unknown3)
	ThermoRead(file, &data.Unknown4)
	ThermoRead(file, &data.Nsegs)
	ThermoRead(file, &data.Unknown5)
	ThermoRead(file, &data.Unknown6)
	ThermoRead(file, &data.Own_addr32)
	ThermoRead(file, &data.Unknown7)
	ThermoRead(file, &data.Unknown8)

	if v >= 64 {
		ThermoRead(file, &data.Scanindex_addr)
		ThermoRead(file, &data.Data_addr)
		ThermoRead(file, &data.Instlog_addr)
		ThermoRead(file, &data.Errorlog_addr)
		ThermoRead(file, &data.Unknown9)
		ThermoRead(file, &data.Scantrailer_addr)
		ThermoRead(file, &data.Scanparams_addr)
		ThermoRead(file, &data.Unknown10)
		ThermoRead(file, &data.Own_addr)

		ThermoRead(file, &data.Unknown11)
		ThermoRead(file, &data.Unknown12)
		ThermoRead(file, &data.Unknown13)
		ThermoRead(file, &data.Unknown14)
		ThermoRead(file, &data.Unknown15)
		ThermoRead(file, &data.Unknown16)
		ThermoRead(file, &data.Unknown17)
		ThermoRead(file, &data.Unknown18)
		ThermoRead(file, &data.Unknown19)
		ThermoRead(file, &data.Unknown20)
		ThermoRead(file, &data.Unknown21)
		ThermoRead(file, &data.Unknown22)
		ThermoRead(file, &data.Unknown23)
		ThermoRead(file, &data.Unknown24)
		ThermoRead(file, &data.Unknown25)
		ThermoRead(file, &data.Unknown26)
		ThermoRead(file, &data.Unknown27)
		ThermoRead(file, &data.Unknown28)
		ThermoRead(file, &data.Unknown29)
		ThermoRead(file, &data.Unknown30)
		ThermoRead(file, &data.Unknown31)
		ThermoRead(file, &data.Unknown32)
		ThermoRead(file, &data.Unknown33)
		ThermoRead(file, &data.Unknown34)
	}

	pos, err = file.Seek(0, 1)
	if err != nil {
		log.Fatal(err)
	}
	return *data, pos
}

func ThermoRead(r io.Reader, data interface{}) {
	switch v := data.(type) {
	case *PascalString:
		binary.Read(r, binary.LittleEndian, &v.Length)
		v.Text = make([]uint16, v.Length)
		binary.Read(r, binary.LittleEndian, &v.Text)
	default:
		binary.Read(r, binary.LittleEndian, v)
	}
}

func ReadInfo(fn string, pos int64, v version) (Info, int64) {
	file, err := os.Open(fn)
	pos, err = file.Seek(pos, 0)
	if err != nil {
		log.Fatal(err)
	}

	data := new(Info)

	ThermoRead(file, &data.Preamble.Methodfilepresent)
	ThermoRead(file, &data.Preamble.Year)
	ThermoRead(file, &data.Preamble.Month)
	ThermoRead(file, &data.Preamble.Weekday)
	ThermoRead(file, &data.Preamble.Day)
	ThermoRead(file, &data.Preamble.Hour)
	ThermoRead(file, &data.Preamble.Minute)
	ThermoRead(file, &data.Preamble.Second)
	ThermoRead(file, &data.Preamble.Millisecond)

	if v >= 57 {
		ThermoRead(file, &data.Preamble.Unknown1)
		ThermoRead(file, &data.Preamble.Data_addr32)
		ThermoRead(file, &data.Preamble.Unknown2)
		ThermoRead(file, &data.Preamble.Unknown3)
		ThermoRead(file, &data.Preamble.Unknown4)
		ThermoRead(file, &data.Preamble.Unknown5)
		ThermoRead(file, &data.Preamble.Runheader_addr32)
		if v <= 63 {
			data.Preamble.Unknown6 = make([]byte, 756)
		} else {
			data.Preamble.Unknown6 = make([]byte, 760)
		}
		ThermoRead(file, &data.Preamble.Unknown6)
	}
	if v >= 64 {
		ThermoRead(file, &data.Preamble.Data_addr)
		ThermoRead(file, &data.Preamble.Unknown7)
		ThermoRead(file, &data.Preamble.Unknown8)
		ThermoRead(file, &data.Preamble.Runheader_addr)

		if v <= 66 {
			data.Preamble.Unknown9 = make([]byte, 1008)
		} else {
			data.Preamble.Unknown9 = make([]byte, 1024)
		}
		ThermoRead(file, &data.Preamble.Unknown9)
	}

	ThermoRead(file, &data.Heading1)
	ThermoRead(file, &data.Heading2)
	ThermoRead(file, &data.Heading3)
	ThermoRead(file, &data.Heading4)
	ThermoRead(file, &data.Heading5)
	ThermoRead(file, &data.Unknown1)

	pos, err = file.Seek(0, 1)
	if err != nil {
		log.Fatal(err)
	}
	return *data, pos
}

func ReadAutoSamplerInfo(fn string, pos int64) (AutoSamplerInfo, int64) {
	file, err := os.Open(fn)
	pos, err = file.Seek(pos, 0)
	if err != nil {
		log.Fatal(err)
	}
	data := new(AutoSamplerInfo)

	ThermoRead(file, &data.Preamble)
	ThermoRead(file, &data.Text)

	pos, err = file.Seek(0, 1)
	if err != nil {
		log.Fatal(err)
	}
	return *data, pos
}

func ReadSequencerRow(fn string, pos int64, v version) (SequencerRow, int64) {
	file, err := os.Open(fn)
	pos, err = file.Seek(pos, 0)
	if err != nil {
		log.Fatal(err)
	}

	data := new(SequencerRow)
	ThermoRead(file, &data.Injection)

	ThermoRead(file, &data.Unknown1)
	ThermoRead(file, &data.Unknown2)
	ThermoRead(file, &data.Id)
	ThermoRead(file, &data.Comment)
	ThermoRead(file, &data.Userlabel1)
	ThermoRead(file, &data.Userlabel2)
	ThermoRead(file, &data.Userlabel3)
	ThermoRead(file, &data.Userlabel4)
	ThermoRead(file, &data.Userlabel5)
	ThermoRead(file, &data.Instmethod)
	ThermoRead(file, &data.Procmethod)
	ThermoRead(file, &data.Filename)
	ThermoRead(file, &data.Path)

	if v >= 57 {
		ThermoRead(file, &data.Vial)
		ThermoRead(file, &data.Unknown3)
		ThermoRead(file, &data.Unknown4)
		ThermoRead(file, &data.Unknown5)
	}
	if v >= 60 {
		ThermoRead(file, &data.Unknown6)
		ThermoRead(file, &data.Unknown7)
		ThermoRead(file, &data.Unknown8)
		ThermoRead(file, &data.Unknown9)
		ThermoRead(file, &data.Unknown10)
		ThermoRead(file, &data.Unknown11)
		ThermoRead(file, &data.Unknown12)
		ThermoRead(file, &data.Unknown13)
		ThermoRead(file, &data.Unknown14)
		ThermoRead(file, &data.Unknown15)
		ThermoRead(file, &data.Unknown16)
		ThermoRead(file, &data.Unknown17)
		ThermoRead(file, &data.Unknown18)
		ThermoRead(file, &data.Unknown19)
		ThermoRead(file, &data.Unknown20)
	}

	pos, err = file.Seek(0, 1)
	if err != nil {
		log.Fatal(err)
	}
	return *data, pos
}

func ReadFileHeader(filename string) (FileHeader, int64) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	data := new(FileHeader)
	err = binary.Read(file, binary.LittleEndian, data)
	if err != nil {
		log.Fatal(err)
	}

	pos, err := file.Seek(0, 1)
	if err != nil {
		log.Fatal(err)
	}
	return *data, pos

}
