package main

import "flag"
import "runtime"
import "time"
import "log"
import "os"
import "getYahooData"
import "strconv"

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	var tend = time.Now()
	var tstart = time.Now().AddDate(0, 0, -150)

	var startDateString = flag.String("startDate", tstart.Format("2006-01-02"), "[Optional] The start date in yyyy-mm-dd format.")
	var endDateString = flag.String("endDate", tend.Format("2006-01-02"), "[Optional] The end date in yyyy-mm-dd format.")
	var numDays = flag.Int("days", 150,
		"[Optional] How many days back to retrieve data from. Takes precedence over Start Date. Default is 150.")

	var numGoroutines = flag.Int("routines", 4,
		"[Optional] Number of Go Routines to be used when retrieving data. Default: 4.")

	var symbolFile = flag.String("symbols", "symbols.txt",
		"[Optional] The file containing the symbols to get data for.")

	flag.Parse()

	symbolFileExists(*symbolFile)

	startDate, err := time.Parse("2006-01-02", *startDateString)
	if err != nil {
		log.Fatal("The Start Date must be provided in the yyyy-mm-dd format.")
	}

	endDate, err := time.Parse("2006-01-02", *endDateString)
	if err != nil {
		log.Fatal("The End Date must be provided in the yyy-mm-dd format.")
	}

	if *numDays != 150 {
		startDate = endDate.AddDate(0, 0, -1*(*numDays))
	}

	getYahooData.NumRoutines = *numGoroutines
	getYahooData.EndDate = endDate
	getYahooData.StartDate = startDate
	getYahooData.SymbolsFile = *symbolFile

	log.Println("Start Date: " + startDate.String())
	log.Println("End Date: " + endDate.String())
	log.Println("Number of Days : " + strconv.Itoa(*numDays))
	log.Println("Number of GoRoutines: " + strconv.Itoa(*numGoroutines))
	log.Println("Symbols file: " + *symbolFile)

	log.Println("Starting")

	getYahooData.Run()

	log.Println("Finished")

}

func symbolFileExists(symbolFile string) {
	file, err := os.Open(symbolFile)
	if err != nil {
		log.Fatal("Cannot find file " + symbolFile)
	}

	file.Close()
}
