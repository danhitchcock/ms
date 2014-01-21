package unthermo

//Only deals with version 57 and above

import (
	"encoding/binary"
	"io"
	"unicode/utf16"
)

/*
 * ScanDataPackets is a list of MS scan packets, containing Centroid Peak
 * or Profile intensities
 */

type ScanDataPackets []ScanDataPacket

func (data ScanDataPackets) Read(r io.Reader, v Version) {
	for i := range data {
		data[i].Read(r, v)
	}
}

func (data *ScanDataPacket) Read(r io.Reader, v Version) {
	Read(r, &data.Header)

	if data.Header.ProfileSize > 0 {
		Read(r, &data.Profile.FirstValue)
		Read(r, &data.Profile.Step)
		Read(r, &data.Profile.PeakCount)
		Read(r, &data.Profile.Nbins)

		data.Profile.Chunks = make([]ProfileChunk, data.Profile.PeakCount)

		for i := uint32(0); i < data.Profile.PeakCount; i++ {
			Read(r, &data.Profile.Chunks[i].Firstbin)
			Read(r, &data.Profile.Chunks[i].Nbins)
			if data.Header.Layout > 0 {
				Read(r, &data.Profile.Chunks[i].Fudge)
			}
			data.Profile.Chunks[i].Signal = make([]float32, data.Profile.Chunks[i].Nbins)
			Read(r, data.Profile.Chunks[i].Signal)
		}
	}

	if data.Header.PeaklistSize > 0 {
		Read(r, &data.PeakList.Count)
		data.PeakList.Peaks = make([]CentroidedPeak, data.PeakList.Count)
		Read(r, data.PeakList.Peaks)
	}

	data.DescriptorList = make([]PeakDescriptor, data.Header.DescriptorListSize)
	Read(r, data.DescriptorList)
	
	data.Unknown = make([]float32, data.Header.UnknownStreamSize)
	Read(r, data.Unknown)
	
	data.Triplets = make([]float32, data.Header.TripletStreamSize)
	Read(r, data.Triplets)

}

type ScanDataPacket struct {
	Header         PacketHeader
	Profile        Profile
	PeakList       PeakList
	DescriptorList []PeakDescriptor
	Unknown        []float32
	Triplets       []float32
}

type PeakDescriptor struct {
	Index  uint16
	Flags  uint8
	Charge uint8
}
type PeakList struct {
	Count uint32
	Peaks []CentroidedPeak
}

type CentroidedPeak struct {
	Mz        float32
	Abundance float32
}

type PacketHeader struct {
	Unknown1           uint32
	ProfileSize        uint32
	PeaklistSize       uint32
	Layout             uint32
	DescriptorListSize uint32
	UnknownStreamSize  uint32
	TripletStreamSize  uint32
	Unknown2           uint32
	Lowmz              float32
	Highmz             float32
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

/*
 * I have currently no idea what TrailerLength is
 */

type TrailerLength uint32

func (data *TrailerLength) Read(r io.Reader, v Version) {
	Read(r, data)
}

/*
 * ScanEvents are encoded headers of the MS scans, their Preamble
 * contain the MS level, type. Events themselves contain range, and
 * conversion parameters from Hz to m/z
 */

type Scanevents []ScanEvent

func (data Scanevents) Read(r io.Reader, v Version) {
	for i := range data {
		data[i].Read(r, v)
	}
}

func (data *ScanEvent) Read(r io.Reader, v Version) {
	if v < 66 {
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
		Read(r, &data.Nprecursors)  //this is just a guess according to Gene Selkov
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

type ScanEvent struct {
	Preamble    [132]uint8 //128 bytes from v63 on, 120 in v62, 80 in v57, 41 below that
	Nprecursors uint32

	Reaction []Reaction

	Unknown1 [13]uint32
	MZrange  [3]FractionCollector
	Nparam   uint32

	Unknown2 [4]float64
	A        float64
	B        float64
	C        float64
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

/*
 * The scan index entries are a list of pointers to the scans
 * other important information is the scan time
 */

type ScanIndexEntries []ScanIndexEntry

func (data ScanIndexEntries) Read(r io.Reader, v Version) {
	for i := range data {
		data[i].Read(r, v)
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
		Read(r, &data.Index) //starts from 0
		Read(r, &data.Scanevent)
		Read(r, &data.Scansegment)
		Read(r, &data.Next)
		Read(r, &data.Unknown1)
		Read(r, &data.DataPacketSize)
		Read(r, &data.Time)
		Read(r, &data.Totalcurrent)
		Read(r, &data.Baseintensity)
		Read(r, &data.Basemz)
		Read(r, &data.Lowmz)
		Read(r, &data.Highmz)
		Read(r, &data.Offset)
	} else {
		Read(r, &data.Offset32)
		Read(r, &data.Index) //starts from 0
		Read(r, &data.Scanevent)
		Read(r, &data.Scansegment)
		Read(r, &data.Next)
		Read(r, &data.Unknown1)
		Read(r, &data.DataPacketSize)
		Read(r, &data.Time)
		Read(r, &data.Totalcurrent)
		Read(r, &data.Baseintensity)
		Read(r, &data.Basemz)
		Read(r, &data.Lowmz)
		Read(r, &data.Highmz)

		data.Offset = uint64(data.Offset32)
	}
}

type ScanIndexEntry struct {
	Offset32       uint32
	Index          uint32
	Scanevent      uint16
	Scansegment    uint16
	Next           uint32
	Unknown1       uint32
	DataPacketSize uint32
	Time           float64
	Totalcurrent   float64
	Baseintensity  float64
	Basemz         float64
	Lowmz          float64
	Highmz         float64
	Offset         uint64
	Unknown2       uint32
	Unknown3       uint32
}

/*
 * Index entries for Chromatography data
 */

type CIndexEntries []CIndexEntry

func (data CIndexEntries) Read(r io.Reader, v Version) {
	for i := range data {
		data[i].Read(r, v)
	}
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

/*
 * CDataPackets are the data from Chromatography machines
 */

type CDataPackets []CDataPacket

func (data CDataPackets) Read(r io.Reader, v Version) {
	for i := range data {
		data[i].Read(r, v)
	}
}

func (data *CDataPacket) Read(r io.Reader, v Version) {
	Read(r, data)
}

type CDataPacket struct { //16 bytes
	Value float64
	Time  float64
}

/*
 * RunHeaders contain all data addresses for data that a certain machine
 * connected to the Mass Spectrometer (including the MS itself)
 * has acquired. Also SN data is available
 */

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

type filename [260]uint16

func (t filename) String() string {
	return string(utf16.Decode(t[:]))
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

/*
 * AutoSamplerInfo comes from the sampling device
 */

func (data *AutoSamplerInfo) Read(r io.Reader, v Version) {
	Read(r, &data.Preamble)
	Read(r, &data.Text)
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

/*
 * SequencerRow contains more information about what the autosampler did
 */

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

/*
 * RawFileInfo contains the addresses of the different RunHeaders,
 * (header of the data that each connected instrument produced)
 * also the acquisition date
 */

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

type RawFileInfo struct {
	Preamble InfoPreamble
	Heading1 PascalString
	Heading2 PascalString
	Heading3 PascalString
	Heading4 PascalString
	Heading5 PascalString
	Unknown1 PascalString
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
	Padding2      [1032]byte //1024 bytes, 1008 bytes in v64
}

type PascalString struct {
	Length int32
	Text   []uint16
}

func (t PascalString) String() string {
	return string(utf16.Decode(t.Text[:]))
}

//Wrapper around binary.Read, reads both PascalStrings and structs from r
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

/*
 * The FileHeaders most valuable piece of info is the file version
 */

func (data *FileHeader) Read(r io.Reader, v Version) {
	Read(r, data)
}

type Version uint32

type FileHeader struct { //1356 bytes
	Magic       uint16    //2 bytes
	Signature   signature //18 bytes
	Unknown1    uint32    //4 bytes
	Unknown2    uint32    //4 bytes
	Unknown3    uint32    //4 bytes
	Unknown4    uint32    //4 bytes
	Version     Version   //4 bytes
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

type headertag [514]uint16

func (t headertag) String() string {
	return string(utf16.Decode(t[:]))
}

type signature [9]uint16

func (t signature) String() string {
	return string(utf16.Decode(t[:]))
}
