package unthermo

import (
	"unicode/utf16"
)

type AutoSamplerInfo struct {
	Preamble AutoSamplerPreamble
	Text     PascalString
}

type AutoSamplerPreamble struct {
	Unknown1      uint32
	Unknown2      uint32
	NumberOfWells uint32
	Unknown3      uint32
	Unknown4      uint32
	Unknown15     uint32
}

type SequencerRow struct {
	Injection  InjectionData
	Unknown1   PascalString
	Unknown2   PascalString
	Id         PascalString
	Comment    PascalString
	Userlabel1 PascalString
	Userlabel2 PascalString
	Userlabel3 PascalString
	Userlabel4 PascalString
	Userlabel5 PascalString
	Instmethod PascalString
	Procmethod PascalString
	Filename   PascalString
	Path       PascalString

	Vial     PascalString
	Unknown3 PascalString
	Unknown4 PascalString
	Unknown5 uint32

	Unknown6  PascalString
	Unknown7  PascalString
	Unknown8  PascalString
	Unknown9  PascalString
	Unknown10 PascalString
	Unknown11 PascalString
	Unknown12 PascalString
	Unknown13 PascalString
	Unknown14 PascalString
	Unknown15 PascalString
	Unknown16 PascalString
	Unknown17 PascalString
	Unknown18 PascalString
	Unknown19 PascalString
	Unknown20 PascalString
}

type InjectionData struct {
	Unknown1                    uint32
	Rownumber                   uint32
	Unknown2                    uint32
	Vial                        [6]uint16 //utf-16
	Injectionvolume             float64
	SampleWeight                float64
	SampleVolume                float64
	InternationalStandardAmount float64
	Dilutionfactor              float64
}

type Info struct {
	Preamble InfoPreamble
	Heading1 PascalString
	Heading2 PascalString
	Heading3 PascalString
	Heading4 PascalString
	Heading5 PascalString
	Unknown1 PascalString
}

type PascalString struct {
	Length int32
	Text   []uint16
}

func (t PascalString) String() string {
	return string(utf16.Decode(t.Text[:]))
}

type InfoPreamble struct {
	Methodfilepresent uint32
	Year              uint16
	Month             uint16
	Weekday           uint16
	Day               uint16
	Hour              uint16
	Minute            uint16
	Second            uint16
	Millisecond       uint16

	Unknown1       uint32
	Data_addr32        uint32
	Unknown2       uint32
	Unknown3       uint32
	Unknown4       uint32
	Unknown5       uint32
	Runheader_addr32        uint32
	Unknown6       []byte //760 bytes, 756 bytes in 57
	
	Data_addr      uint64
	Unknown7       uint32
	Unknown8       uint32
	Runheader_addr uint64
	Unknown9       []byte //1024 bytes, 1008 bytes in 64
}

type headertag [514]uint16

func (t headertag) String() string {
	return string(utf16.Decode(t[:]))
}

type signature [9]uint16

func (t signature) String() string {
	return string(utf16.Decode(t[:]))
}

type version uint32;

type FileHeader struct { //1356 bytes
	Magic       uint16    //2 bytes
	Signature   signature //18 bytes
	Unknown1    uint32    //4 bytes
	Unknown2    uint32    //4 bytes
	Unknown3    uint32    //4 bytes
	Unknown4    uint32    //4 bytes
	Version     version    //4 bytes
	Audit_start AuditTag  //112 bytes
	Audit_end   AuditTag  //112 bytes
	Unknown5    uint32    //4 bytes
	Unknown6    [60]byte  //60 bytes
	Tag         headertag //1028 bytes
}

type audittag [25]uint16

func (t audittag) String() string {
	return string(utf16.Decode(t[:]))
}

type AuditTag struct { //112 bytes
	Time     uint64   //8 bytes Windows 64-bit timestamp
	Tag_1    audittag //50 bytes
	Tag_2    audittag
	Unknown1 uint32 //4 bytes
}
