package getYahooData

import "encoding/csv"
import "os"
import "fmt"
import "net/http"
import "time"
import "log"

// read symbols from the file
func readSymbols(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal("Error:", err)
	}

	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Error:", err)
	}

	var list []string

	for _, record := range records {
		list = append(list, record[0])
	}

	return list
}

func getYahooData(symbol string) {

	log.Printf("GETTING %v\n", fmt.Sprintf(yahooRequestString, symbol, startMonth-1, startDay, startYear, endMonth-1, endDay, endYear))

	response, err := http.Get(fmt.Sprintf(yahooRequestString, symbol, startMonth-1, startDay, startYear, endMonth-1, endDay, endYear))
	if err != nil {
		log.Printf("Error %v\n", err)
		return
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Printf("WARNING: Yahoo could not find data for symbol %v\n", symbol)
		return
	}

	csvReader := csv.NewReader(response.Body)

	records, err := csvReader.ReadAll()

	if err != nil {
		log.Printf("Error %v\n", err)
		return
	}

	// reverse the order since yahoo gives us descending date order
	// but not the first row since it's the header
	for i, j := 1, len(records)-1; i < j; i, j = i+1, j-1 {
		records[i], records[j] = records[j], records[i]
	}

	file, err := os.Create(symbol + ".csv")
	if err != nil {
		log.Printf("Error %v\n", err)
		return
	}

	defer file.Close()

	csvWriter := csv.NewWriter(file)

	csvWriter.WriteAll(records)
}

func pullData(coms chan string, doneSignal chan bool, routineNum int) {
	for {
		symbol := <-coms

		log.Printf("Routine %v will retrieve symbol %v\n", routineNum, symbol)

		if symbol == endMarker {
			doneSignal <- true
			return
		}

		getYahooData(symbol)

		// sleep for 5 seconds so we don't get throttled
		time.Sleep(5 * time.Second)
	}
}

func pushSymbols(coms chan string, list []string) {
	for _, symbol := range list {
		coms <- symbol
	}

	for i := 0; i < NumRoutines; i++ {
		coms <- endMarker
	}
}

var endMarker = "^THEEND^"

var startYear = 2012
var startMonth = 1
var startDay = 1
var endYear = 2012
var endMonth = 11
var endDay = 11

// get daily data from 2011-1-1 to 2011-12-31
var yahooRequestString = "http://ichart.yahoo.com/table.csv?s=%v&a=%v&b=%v&c=%v&d=%v&e=%v&f=%v&g=d&ignore=.csv"

var SymbolsFile string
var NumRoutines int

var StartDate time.Time
var EndDate time.Time

func Run() {

	// read list of symbols
	symbols := readSymbols(SymbolsFile)

	if symbols == nil {
		log.Fatal("Error: empty symbols file")
	}

	// make communication channel
	communication := make(chan string, NumRoutines)
	done := make(chan bool, 1)

	var month time.Month
	// set the start and end dates
	endYear, month, endDay = EndDate.Date()
	endMonth = int(month)

	startYear, month, startDay = StartDate.Date()
	startMonth = int(month)

	// create appropriate folder
	var folderName = fmt.Sprintf("%v_%v_%v", endYear, endMonth, endDay)
	os.Mkdir(folderName, os.ModeDir)

	os.Chdir(folderName)

	// start up pull threads
	for i := 0; i < NumRoutines; i++ {
		go pullData(communication, done, i)
	}

	// start push thread
	go pushSymbols(communication, symbols)

	for i := 0; i < NumRoutines; i++ {
		<-done
	}
}
