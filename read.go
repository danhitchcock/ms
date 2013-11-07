package unthermo

import (
	"encoding/binary"
	"io"
	"log"
	"os"
)

func ReadInfo(fn string, pos int64, v version) (Info, int64) {
	file, err := os.Open(fn)
	pos, err = file.Seek(pos, 0)
	if err != nil {
		log.Fatal(err)
	}

	data := new(Info)


	binary.Read(file, binary.LittleEndian, &data.Preamble.Methodfilepresent)
	binary.Read(file, binary.LittleEndian, &data.Preamble.Year)
	binary.Read(file, binary.LittleEndian, &data.Preamble.Month)
	binary.Read(file, binary.LittleEndian, &data.Preamble.Weekday)
	binary.Read(file, binary.LittleEndian, &data.Preamble.Day)
	binary.Read(file, binary.LittleEndian, &data.Preamble.Hour)
	binary.Read(file, binary.LittleEndian, &data.Preamble.Minute)
	binary.Read(file, binary.LittleEndian, &data.Preamble.Second)
	binary.Read(file, binary.LittleEndian, &data.Preamble.Millisecond)

	if v >= 57 {
		binary.Read(file, binary.LittleEndian, &data.Preamble.Unknown1)
		binary.Read(file, binary.LittleEndian, &data.Preamble.Data_addr32)
		binary.Read(file, binary.LittleEndian, &data.Preamble.Unknown2)
		binary.Read(file, binary.LittleEndian, &data.Preamble.Unknown3)
		binary.Read(file, binary.LittleEndian, &data.Preamble.Unknown4)
		binary.Read(file, binary.LittleEndian, &data.Preamble.Unknown5)
		binary.Read(file, binary.LittleEndian, &data.Preamble.Runheader_addr32)
		if v <= 63 {
			data.Preamble.Unknown6 = make([]byte, 756)
		} else {
			data.Preamble.Unknown6 = make([]byte, 760)
		}
		binary.Read(file, binary.LittleEndian, &data.Preamble.Unknown6)
	}
	if v >= 64 {
		binary.Read(file, binary.LittleEndian, &data.Preamble.Data_addr)
		binary.Read(file, binary.LittleEndian, &data.Preamble.Unknown7)
		binary.Read(file, binary.LittleEndian, &data.Preamble.Unknown8)
		binary.Read(file, binary.LittleEndian, &data.Preamble.Runheader_addr)

		if v <= 66 {
			data.Preamble.Unknown9 = make([]byte, 1008)
		} else {
			data.Preamble.Unknown9 = make([]byte, 1024)
		}
		binary.Read(file, binary.LittleEndian, &data.Preamble.Unknown9)
	}

	data.Heading1 = ReadPascalString(file)
	data.Heading2 = ReadPascalString(file)
	data.Heading3 = ReadPascalString(file)
	data.Heading4 = ReadPascalString(file)
	data.Heading5 = ReadPascalString(file)
	data.Unknown1 = ReadPascalString(file)

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
	err = binary.Read(file, binary.LittleEndian, &data.Preamble)
	if err != nil {
		log.Fatal(err)
	}

	data.Text = ReadPascalString(file)

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
	err = binary.Read(file, binary.LittleEndian, &data.Injection)
	if err != nil {
		log.Fatal(err)
	}

	data.Unknown1 = ReadPascalString(file)
	data.Unknown2 = ReadPascalString(file)
	data.Id = ReadPascalString(file)
	data.Comment = ReadPascalString(file)
	data.Userlabel1 = ReadPascalString(file)
	data.Userlabel2 = ReadPascalString(file)
	data.Userlabel3 = ReadPascalString(file)
	data.Userlabel4 = ReadPascalString(file)
	data.Userlabel5 = ReadPascalString(file)
	data.Instmethod = ReadPascalString(file)
	data.Procmethod = ReadPascalString(file)
	data.Filename = ReadPascalString(file)
	data.Path = ReadPascalString(file)

	if v >= 57 {
		data.Vial = ReadPascalString(file)
		data.Unknown3 = ReadPascalString(file)
		data.Unknown4 = ReadPascalString(file)
		binary.Read(file, binary.LittleEndian, &data.Unknown5)
	}
	if v >= 60 {
		data.Unknown6 = ReadPascalString(file)
		data.Unknown7 = ReadPascalString(file)
		data.Unknown8 = ReadPascalString(file)
		data.Unknown9 = ReadPascalString(file)
		data.Unknown10 = ReadPascalString(file)
		data.Unknown11 = ReadPascalString(file)
		data.Unknown12 = ReadPascalString(file)
		data.Unknown13 = ReadPascalString(file)
		data.Unknown14 = ReadPascalString(file)
		data.Unknown15 = ReadPascalString(file)
		data.Unknown16 = ReadPascalString(file)
		data.Unknown17 = ReadPascalString(file)
		data.Unknown18 = ReadPascalString(file)
		data.Unknown19 = ReadPascalString(file)
		data.Unknown20 = ReadPascalString(file)
	}

	pos, err = file.Seek(0, 1)
	if err != nil {
		log.Fatal(err)
	}
	return *data, pos
}

func ReadPascalString(r io.Reader) PascalString {
	data := new(PascalString)
	binary.Read(r, binary.LittleEndian, &data.Length)

	data.Text = make([]uint16, data.Length)

	binary.Read(r, binary.LittleEndian, &data.Text)

	return *data
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
