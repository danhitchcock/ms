package unthermo

//Only deals with version 57 and above

import (
	"unicode/utf16"
)

type PacketHeader struct {
	Unknown1            uint32
	Profilesize         uint32
	Peaklistsize        uint32
	Layout              uint32
	Descriptorlistsize  uint32
	Sizeofunknownstream uint32
	Sizeoftripletstream uint32
	Unknown2            uint32
	Lowmz               float32
	Highmz              float32
}

type ScanIndexEntry struct {
	Offset32      uint32
	Index         uint16
	Scanevent     uint16
	Scansegment   uint16
	Next          uint32
	Unknown1      uint32
	Datasize      uint32
	Starttime     float64
	Totalcurrent  float64
	Baseintensity float64
	Basemz        float64
	Lowmz         float64
	Highmz        float64
	Offset        uint64
}

type filename [260]uint16

func (t filename) String() string {
	return string(utf16.Decode(t[:]))
}

type RunHeader struct {
	SampleInfo         SampleInfo
	Filename1          filename
	Filename2          filename
	Filename3          filename
	Filename4          filename
	Filename5          filename
	Filename6          filename
	Unknown1           float64
	Unknown2           float64
	Filename7          filename
	Filename8          filename
	Filename9          filename
	Filename10         filename
	Filename11         filename
	Filename12         filename
	Filename13         filename
	Scantrailer_addr32 uint32
	Scanparams_addr32  uint32
	Unknown3           uint32
	Unknown4           uint32
	Nsegs              uint32
	Unknown5           uint32
	Unknown6           uint32
	Own_addr32         uint32
	Unknown7           uint32
	Unknown8           uint32

	Scanindex_addr   uint64
	Data_addr        uint64
	Instlog_addr     uint64
	Errorlog_addr    uint64
	Unknown9         uint64
	Scantrailer_addr uint64
	Scanparams_addr  uint64
	Unknown10        uint64
	Own_addr         uint64

	Unknown11 uint32
	Unknown12 uint32
	Unknown13 uint32
	Unknown14 uint32
	Unknown15 uint32
	Unknown16 uint32
	Unknown17 uint32
	Unknown18 uint32
	Unknown19 uint32
	Unknown20 uint32
	Unknown21 uint32
	Unknown22 uint32
	Unknown23 uint32
	Unknown24 uint32
	Unknown25 uint32
	Unknown26 uint32
	Unknown27 uint32
	Unknown28 uint32
	Unknown29 uint32
	Unknown30 uint32
	Unknown31 uint32
	Unknown32 uint32
	Unknown33 uint32
	Unknown34 uint32
}

type SampleInfo struct {
	Unknown1         uint32
	Unknown2         uint32
	FirstScanNumber  uint32
	LastScanNumber   uint32
	InstLogLength    uint32
	Unknown3         uint32
	Unknown4         uint32
	ScanIndexAddress uint32
	DataAddress      uint32
	InstLogAddress   uint32
	ErrorLogAddress  uint32
	Unknown5         uint32
	MaxIonCurrent    float64
	Lowmz            float64
	Highmz           float64
	Starttime        float64
	Endtime          float64
	Unknown6         [56]byte
	Tag1             [44]uint16
	Tag2             [20]uint16
	Tag3             [160]uint16
}

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

	Unknown1         uint32
	Data_addr32      uint32
	Unknown2         uint32
	Unknown3         uint32
	Unknown4         uint32
	Unknown5         uint32
	Runheader_addr32 uint32
	Unknown6         []byte //760 bytes, 756 bytes in v57

	Data_addr      uint64
	Unknown7       uint32
	Unknown8       uint32
	Runheader_addr uint64
	Unknown9       []byte //1024 bytes, 1008 bytes in v64
}

type headertag [514]uint16

func (t headertag) String() string {
	return string(utf16.Decode(t[:]))
}

type signature [9]uint16

func (t signature) String() string {
	return string(utf16.Decode(t[:]))
}

type version uint32

type FileHeader struct { //1356 bytes
	Magic       uint16    //2 bytes
	Signature   signature //18 bytes
	Unknown1    uint32    //4 bytes
	Unknown2    uint32    //4 bytes
	Unknown3    uint32    //4 bytes
	Unknown4    uint32    //4 bytes
	Version     version   //4 bytes
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
