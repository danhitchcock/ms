//Package UnThermo
//after unfinnigan http://code.google.com/p/unfinnigan/wiki/FileLayoutOverview

//example use case:
package main

import (
	"bitbucket.org/proteinspector/unthermo"
	"flag"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sort"
)

func main() {
	var scan uint64
	var instr int
	var mz mzarg
	var tol float64
	var proc bool
	flag.Uint64Var(&scan, "sc", 0, "scan number")
	flag.IntVar(&instr, "instr", 0, "instrument number (0 is the mass spec, 1 or higher can be other attached controllers such as an LC)")
	flag.Var(&mz, "mz", "m/z to filter on, can also be a range by specifying the flag twice (or in the format min-max)")
	flag.Float64Var(&tol, "tol", 0, "allowed mz tolerance, can be used with -mz")
	flag.BoolVar(&proc, "p", false, "process the mass list")
	flag.Parse()

	for _, filename := range flag.Args() {
		ch := make(chan MS)
		if scan > 0 && instr == 0 {
			go ReadScan(filename, scan, ch)
		} else if instr > 0 {
			PrintOther(filename, instr)
		} else {
			go ReadAllScans(filename, ch)
		}
		if !proc { //just print
			switch len(mz) {
				case 0:
					PrintMZI(ch, 0, math.MaxFloat64)
				case 1:
					PrintMZI(ch, mz[0]-tol, mz[0]+tol)
				case 2:
					if mz[0]<=mz[1] {
						PrintMZI(ch, mz[0]-tol, mz[1]+tol)
					}
			}
		} else {
			Process(ch)
		}
	}
}

func Process(in <-chan MS) {
	//discovery: (instead of 2 times searching, search as many times as needed. kthxbai)
	//first sort on intensity, then for every event,
	//look for more with m/z's close to it (pick peaks) (of course neighboring scans first(how many))
	//then browse for isotope peaks at (expected(close for now?)) intensities to determine charge
	//for the possible molecules (charge,mass) look for relatives
	//propagate the intermediate results further into the pipeline,
	//the succesive pieces each deal with more specific peptide information for a search
		var m []MS
		for ms := range in {
			if ms.mz <= 723 && ms.mz >= 720 {
				m = append(m,ms)
			}
		}
		
		ascMz := func(p1, p2 *MS) bool {
			return p1.mz < p2.mz
		}
		descInt := func(p1, p2 *MS) bool {
			return p1.I > p2.I
		}
		By(descInt).Sort(m)
		fmt.Println(m)
		
		By(ascMz).Sort(m)
		fmt.Println(m)
		
}

func PrintMZI(in <-chan MS, lo float64, hi float64) {
	for ms := range in {
		if ms.mz <= hi && ms.mz >= lo {
			fmt.Printf("%6.6f %6.6f", ms.mz, ms.I)
		}
	}
}

func ReadScan(fn string, sn uint64, out chan<- MS) {
	info, ver := unthermo.ReadFileHeaders(fn)

	rh := new(unthermo.RunHeader)
	unthermo.ReadFile(fn, info.Preamble.RunHeaderAddr[0], ver, rh)

	//the MS RunHeader contains besides general info three interesting
	//addresses: ScanindexAddr (with the scan headers), DataAddr,
	//and ScantrailerAddr (which includes orbitrap Hz-m/z conversion
	//parameters and info about the scans)

	if sn < uint64(rh.SampleInfo.FirstScanNumber) || sn > uint64(rh.SampleInfo.LastScanNumber) {
		log.Fatal("scan number out of range: ", rh.SampleInfo.FirstScanNumber, ", ", rh.SampleInfo.LastScanNumber)
	}

	//read the n'th ScanIndexEntry
	sie := new(unthermo.ScanIndexEntry)
	unthermo.ReadFile(fn, rh.ScanindexAddr+(sn-1)*sie.Size(ver), ver, sie)

	//For later conversion of frequency values to m/z, we need a ScanEvent
	//The list of them starts 4 bytes later than ScantrailerAddr
	pos := rh.ScantrailerAddr + 4

	//the ScanEvents are of variable size and have no pointer to
	//them, we need to read at least all the ones preceding n
	scanevent := new(unthermo.ScanEvent)
	for i := uint64(0); i < sn; i++ {
		pos = unthermo.ReadFile(fn, pos, ver, scanevent)
	}

	//read Scan Packet for the above scan number
	scan := new(unthermo.ScanDataPacket)
	unthermo.ReadFile(fn, rh.DataAddr+sie.Offset, 0, scan)

	//convert the Hz values into m/z and list the signals
	for i := uint32(0); i < scan.Profile.PeakCount; i++ {
		for j := uint32(0); j < scan.Profile.Chunks[i].Nbins; j++ {
			out <- MS{sn, uint64(i), scanevent.Convert(scan.Profile.FirstValue+float64(scan.Profile.Chunks[i].Firstbin+j)*scan.Profile.Step) + float64(scan.Profile.Chunks[i].Fudge), scan.Profile.Chunks[i].Signal[j]}
		}
	}
	close(out)
}

func ReadAllScans(fn string, out chan<- MS) {
	//Read necessary headers
	info, ver := unthermo.ReadFileHeaders(fn)
	rh := new(unthermo.RunHeader)
	unthermo.ReadFile(fn, info.Preamble.RunHeaderAddr[0], ver, rh)

	//For later conversion of frequency values to m/z, we need a ScanEvent
	//for each Scan.
	//The list of them starts an uint32 later than ScantrailerAddr
	nScans := uint64(rh.SampleInfo.LastScanNumber - rh.SampleInfo.FirstScanNumber + 1)
	scanevents := make(unthermo.Scanevents, nScans)
	unthermo.ReadFileRange(fn, rh.ScantrailerAddr + 4, rh.ScanparamsAddr, ver, scanevents)
	
	//read all scanindexentries at once, this is probably the fastest
	scanindexentries := make(unthermo.ScanIndexEntries, nScans)
	unthermo.ReadFileRange(fn, rh.ScanindexAddr, rh.ScantrailerAddr, ver, scanindexentries)
	
	offset := make(chan uint64)
	scans := make(chan *unthermo.ScanDataPacket)
	go unthermo.ReadScansFromMemory(fn, rh.DataAddr, rh.OwnAddr, 0, offset, scans)
	
	//for s:=uint64(0); s<nScans; s++ {
	for k,s := range scanindexentries {
		offset <- s.Offset
		scan := <-scans
		//convert the Hz values into m/z and list the signals
		for i := uint32(0); i < scan.Profile.PeakCount; i++ {
			for j := uint32(0); j < scan.Profile.Chunks[i].Nbins; j++ {
				//m=append(m, MS{Scan{uint64(s+1), uint64(i)},scanevents[s].Convert(scan.Profile.FirstValue+float64(scan.Profile.Chunks[i].Firstbin+j)*scan.Profile.Step)+float64(scan.Profile.Chunks[i].Fudge), scan.Profile.Chunks[i].Signal[j]})
				out<-MS{scan: uint64(k+1), packet: uint64(i), mz: scanevents[k].Convert(scan.Profile.FirstValue+float64(scan.Profile.Chunks[i].Firstbin+j)*scan.Profile.Step)+float64(scan.Profile.Chunks[i].Fudge), I: scan.Profile.Chunks[i].Signal[j]}
			}
		}
	}
	close(out)
}

//@pre instr>0. in other words: not the mass spectrometer
func PrintOther(fn string, instr int) {
	info, ver := unthermo.ReadFileHeaders(fn)

	if uint32(instr) > info.Preamble.NControllers-1 {
		log.Fatal(instr, " is higher than number of extra controllers: ", info.Preamble.NControllers-1)
	}

	rh := new(unthermo.RunHeader)
	unthermo.ReadFile(fn, info.Preamble.RunHeaderAddr[instr], ver, rh)

	//The instrument RunHeader contains an interesting address: DataAddr
	//There is another address ScanIndexAddr, which points to CIndexEntry
	//containers at ScanIndexAddr. Less data can be read for now

	nScan := uint64(rh.SampleInfo.LastScanNumber - rh.SampleInfo.FirstScanNumber + 1)
	for i := uint64(1); i < nScan; i++ {
		cdata := new(unthermo.CDataPacket)
		unthermo.ReadFile(fn, rh.DataAddr+i*16, ver, cdata) //16 bytes of CDataPacket
		fmt.Println(cdata.Time, cdata.Value)
	}
}

//type declarations

type MS struct {
	scan     uint64
	packet uint64
	mz   float64
	I    float32
}

type By func(p1, p2 *MS) bool

type msSorter struct {
	mses []MS
	by      By
}

type mzarg []float64 //a new type for passing flags on the command line

//sort functions
func (by By) Sort(mses []MS) {
	ps := &msSorter{
		mses: mses,
		by:      by,
		}
	sort.Sort(ps)
}

func (s *msSorter) Len() int {
	return len(s.mses)
}

func (s *msSorter) Swap(i, j int) {
	s.mses[i], s.mses[j] = s.mses[j], s.mses[i]
}

func (s *msSorter) Less(i, j int) bool {
	return s.by(&s.mses[i], &s.mses[j])
}

//argument parsing
func (i *mzarg) String() string {
	return fmt.Sprintf("%d", *i)
}

func (i *mzarg) Set(value string) error {
	for _, splv := range strings.Split(value, "-") {
		tmp, err := strconv.ParseFloat(splv,64)
		if err != nil {
			return err
		}
		*i = append(*i, tmp)
	}
    return nil
}

