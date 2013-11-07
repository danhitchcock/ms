package unthermo

import (
	"encoding/binary"
	"io"
	"log"
	"os"
)

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

	log.Print(2)
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
