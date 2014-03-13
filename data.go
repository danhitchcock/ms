package ms

type Peak struct {
	Time float64
	Mz   float64
	I    float32
}

type Spectrum []Peak

type Scan struct {
	Spectrum
	MSLevel    uint8
	Activation Activation
}

type Activation int

const (
	CID Activation = iota
	HCD
)

//Spectrum implements sort.Interface for []Peak based on m/z
func (a Spectrum) Len() int           { return len(a) }
func (a Spectrum) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Spectrum) Less(i, j int) bool { return a[i].Mz < a[j].Mz }
