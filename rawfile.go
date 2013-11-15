package unthermo

//Only deals with version 57 and above

import (
	"unicode/utf16"
)

type ScanDataPacket struct {
	Header   PacketHeader
	Profile  Profile
	PeakList PeakList
}

type PeakList struct {
	Count uint32
	Peaks []Peak
}

type Peak struct {
	Mz        float32
	Abundance float32
}

type PacketHeader struct {
	Unknown1            uint32
	ProfileSize         uint32
	PeaklistSize        uint32
	Layout              uint32
	Descriptorlistsize  uint32
	Sizeofunknownstream uint32
	Sizeoftripletstream uint32
	Unknown2            uint32
	Lowmz               float32
	Highmz              float32
}

type Profile struct {
	FirstValue float64
	Step       float64
	PeakCount  uint32
	Nbins      uint32
	Chunks     []ProfileChunk
}

type ProfileChunk struct {
	Firstbin uint32
	Nbins    uint32
	Fudge    float32
	Signal   []float32
}

type ScanEvent struct {
	Preamble    ScanEventPreamble
	Nprecursors uint32

	Reaction []Reaction

	Unknown1 uint32
	MZrange  FractionCollector
	Nparam   uint32

	Unknown2 float64
	A        float64
	B        float64
	C        float64
	D        float64
	E        float64
	I        float64

	Unknown3 uint32
	Unknown4 uint32
}

type Reaction struct {
	Precursormz float64
	Unknown1    float64
	Energy      float64
	Unknown2    uint32
	Unknown3    uint32
}

type FractionCollector struct {
	Lowmz  float64
	Highmz float64
}

type ScanEventPreamble [136]uint8 //128 bytes in v63 and up, 120 in v62, 80 in v57, 41 below that

type ScanIndexEntry struct {
	Offset32       uint32
	Index          uint32
	Scanevent      uint16
	Scansegment    uint16
	Next           uint32
	Unknown1       uint32
	DataPacketSize uint32
	ScanTime       float64
	Totalcurrent   float64
	Baseintensity  float64
	Basemz         float64
	Lowmz          float64
	Highmz         float64
	Offset         uint64
	Unknown2       uint32
	Unknown3       uint32
}

type CIndexEntry struct {
	Offset32 uint32
	Index    uint32
	Event    uint16
	Unknown1 uint16
	Unknown2 uint32
	Unknown3 uint32
	Unknown4 uint32
	Unknown5 float64
	Time     float64
	Unknown6 float64
	Unknown7 float64
	Value    float64

	Offset uint64
}

type CDataPacket struct { //unused at the moment
	Value float64
	Time  float64
}

type filename [260]uint16

func (t filename) String() string {
	return string(utf16.Decode(t[:]))
}

type RunHeader struct {
	SampleInfo        SampleInfo
	Filename1         filename
	Filename2         filename
	Filename3         filename
	Filename4         filename
	Filename5         filename
	Filename6         filename
	Unknown1          float64
	Unknown2          float64
	Filename7         filename
	Filename8         filename
	Filename9         filename
	Filename10        filename
	Filename11        filename
	Filename12        filename
	Filename13        filename
	ScantrailerAddr32 uint32
	ScanparamsAddr32  uint32
	Unknown3          uint32
	Unknown4          uint32
	Nsegs             uint32
	Unknown5          uint32
	Unknown6          uint32
	OwnAddr32         uint32
	Unknown7          uint32
	Unknown8          uint32

	ScanindexAddr   uint64
	DataAddr        uint64
	InstlogAddr     uint64
	ErrorlogAddr    uint64
	Unknown9        uint64
	ScantrailerAddr uint64
	ScanparamsAddr  uint64
	Unknown10       uint32
	Unknown11       uint32
	OwnAddr         uint64

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
	Unknown35 uint32

	Unknown36 [8]byte
	Unknown37 uint32
	Device    PascalString
	Model     PascalString
	SN        PascalString
	SWVer     PascalString
	Tag1      PascalString
	Tag2      PascalString
	Tag3      PascalString
	Tag4      PascalString
}

type SampleInfo struct {
	Unknown1        uint32
	Unknown2        uint32
	FirstScanNumber uint32
	LastScanNumber  uint32
	InstlogLength   uint32
	Unknown3        uint32
	Unknown4        uint32
	ScanindexAddr   uint32 //unused in 64-bit versions
	DataAddr        uint32
	InstlogAddr     uint32
	ErrorlogAddr    uint32
	Unknown5        uint32
	MaxSignal       float64
	Lowmz           float64
	Highmz          float64
	Starttime       float64
	Endtime         float64
	Unknown6        [56]byte
	Tag1            [44]uint16
	Tag2            [20]uint16
	Tag3            [160]uint16
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

type RawFileInfo struct {
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

	Unknown1        uint32
	DataAddr32      uint32
	NControllers    uint32
	NControllers2   uint32
	Unknown2        uint32
	Unknown3        uint32
	RunHeaderAddr32 []uint32
	Unknown4        []uint32
	Unknown5        []uint32
	Padding1        [764]byte //760 bytes, 756 bytes in v57

	DataAddr      uint64
	Unknown6      uint64
	RunHeaderAddr []uint64
	Unknown7      []uint64
	Padding2      [1024]byte //1024 bytes, 1008 bytes in v64
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
