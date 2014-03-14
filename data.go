//Package ms is a library for mass spectrometry data
package ms

//Peak represents an ion peak
type Peak struct {
	Mz float64
	I  float32
}

//Spectrum is the collection of peaks
type Spectrum []Peak

//Scan represents the peak acquisition event of the mass spectrometer
type Scan struct {
	Activation Activation
	Analyzer Analyzer
	MSLevel    uint8
	Spectrum
	Time float64
}

//Activation is a type that describes the activation type for fragmentation scans
type Activation int

const (
	//HCD for Higher-energy collisional dissociation
	HCD Activation = iota + 1
	_
	_
	//CID stands for Collision induced dissociation
	CID
)

type Analyzer int

const (
	ITMS Analyzer = iota
	TQMS
	SQMS
	TOFMS
	FTMS
	Sector
	Undefined
)

//Spectrum implements sort.Interface for []Peak based on m/z

func (a Spectrum) Len() int           { return len(a) }
func (a Spectrum) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Spectrum) Less(i, j int) bool { return a[i].Mz < a[j].Mz }
