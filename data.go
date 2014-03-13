//Package ms is a library for mass spectrometry data
package ms

//An ion peak
type Peak struct {
	Mz   float64
	I    float32
}

//A collection of peaks
type Spectrum []Peak

//The scan event of the mass spectrometer
type Scan struct {
	Activation Activation
	MSLevel    uint8
	Spectrum
	Time float64
}

//For fragmentation scans, there is an activation type
type Activation int

const (
	//Collision induced dissociation
	CID Activation = iota
	//Higher-energy collisional dissociation
	HCD
)

//Spectrum implements sort.Interface for []Peak based on m/z

func (a Spectrum) Len() int           { return len(a) }
func (a Spectrum) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Spectrum) Less(i, j int) bool { return a[i].Mz < a[j].Mz }
