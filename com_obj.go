package main

import (
	"fmt"
	"os"
	"log"
	//"time"
	"encoding/csv"
	"strings"
	"io/ioutil"
	"os/exec"
	//"golang.org/x/tools/godoc"
	"sync"
	"runtime"
	"flag"
	"runtime/pprof"
	"stat_wraper/int_libs"
	"github.com/hashicorp/logutils"
	"github.com/xuri/excelize"
	//"github.com/tealeg/xlsx"
	"strconv"
)

var ver_string = "version = 0.3.0, build date = 2017-10-20"

// TODO: flags for different concurrency modes

var cwd, _ = os.Getwd()
var reports_dir = cwd + "\\REPORTS\\"
var scripts_dir  = cwd + "\\SCRIPTS\\"
var csv_dir = cwd + "\\CSV\\"
var stat_dir = cwd + "\\STAT\\"
var debug_levels = []string{"DEBUG", "INFO", "WARN", "ERROR"}

var sheets_ids = map[int]string{1: "graph", 2: "2G", 3: "3G"}
var sheets_names = map[string]int{"graph": 1, "2G": 2, "3G": 3}

func check(e error) {
	if e != nil {
		log.Panicln("[ERROR]", "error=", e)
	}
}

func set_site_name() string{
	reports_files, _ := ioutil.ReadDir(reports_dir)
	var site_name string
	if len(reports_files) == 0 {
		fmt.Println("previous reports not found in \"REPORTS\" dir. \n" +
			"Need to know SiteName. (for example: KIE1)")
		fmt.Print("SiteName: ")
		fmt.Scanf("%s", &site_name)
	}
	return site_name
}



// creates new report based on template or previous report
func ReportCreator(site_name string) {
	reports_files, _ := ioutil.ReadDir(reports_dir)
	file_name_prefix := "KPI-SGSN_"
	if len(reports_files) == 0 {
		file_name_prefix = "KPI-SGSN_" + site_name + "_"
		log.Println("[DEBUG]", "opening file=", scripts_dir + "SGSN_template.xlsx")
		templateFile, err := excelize.OpenFile(scripts_dir + "SGSN_template.xlsx")
		check(err)
		templateFile.SetCellValue("graph", "J1", site_name)
		datetime := XlsxPrepare(templateFile, 2)
		last_report_file := file_name_prefix + datetime + ".xlsx"
		templateFile.SaveAs(reports_dir + last_report_file)
	} else {
		// DONE: ~$KPI-SGSN_KIE#1_20170608_0600.xlsx ignore such files
		var allowed_files  []os.FileInfo
		for _, file := range reports_files {
			if  strings.HasPrefix(file.Name(), file_name_prefix){
				allowed_files = append(allowed_files, file)
			} else {
				log.Println("[WARN]", "found not allowed filename=", file.Name(), "in REPORTS")
			}
		}
		if len(allowed_files) == 0 {
			log.Println("[ERROR]", "no allowed files to edit in folder:", reports_dir)
			os.Exit(1)
		}
		prev_report_file := allowed_files[len(allowed_files) -1 ].Name()
		log.Println("[INFO]", "last report file with index = -1", prev_report_file)

		templateFile, err := excelize.OpenFile(reports_dir + prev_report_file)
		check(err)
		site_name := templateFile.GetCellValue("graph", "J1")
		log.Println("[DEBUG]", "site_name=", site_name)
		file_name_prefix = "KPI-SGSN_" + site_name + "_"
		//last_row := templateFile.GetRows("2G")
		start_row := 2

		for templateFile.GetCellValue("2G", "A"+strconv.Itoa(start_row)) != "" {
			start_row += 1
		}

		log.Println("[INFO]", "start_row", start_row)
		datetime := XlsxPrepare(templateFile, start_row)
		last_report_file := file_name_prefix + datetime + ".xlsx"
		if last_report_file == prev_report_file{
			log.Println("[ERROR]", "trying to overite exist file (posible reason: reused data in CSV or STAT dir)")
			os.Exit(1)
		}
		templateFile.SaveAs(reports_dir + last_report_file)
		log.Println("[INFO]", "SaveAs:", last_report_file)
	}
}

// call perl script
func call_script(profile string)  {
	profile = "--profile=.\\SCRIPTS\\" + profile
	log.Println("[INFO]", "calling PERL script")
	cmd := exec.Command("cmd", "/C", ".\\SCRIPTS\\wxp.pl", "--file=STAT", "--outdir=CSV", "--csv",
		profile, "--nosilent", "--service=safmm")
	//cmd := exec.Command("cmd", "/C", "date", "/T")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)

	}
	log.Println("[INFO]", "end PERL script")
}

func read_csv_file(file_name string) (headers []string, data [][]string) {
	//data, err := ioutil.ReadFile(file_name)
	f, err := os.Open(file_name)
	check(err)
	defer f.Close()
	//data_io := io.Reader(data)
	//r := csv.NewReader(bufio.NewReader(f))
	r := csv.NewReader(f)
	r.Comma = ';'
	records, err := r.ReadAll()
	headers = records[0]
	data = records[1:]
	return
}

func XlsxPrepare(xlsx_in *excelize.File, start_row int) (date_time string) {
	// modify xlsx stucture based on CSV
	file_s_csv, _ := ioutil.ReadDir(csv_dir)
	date := ""
	time := ""
	for _, csv_file := range file_s_csv {
		headers, data := read_csv_file(csv_dir + csv_file.Name())

		var sheet string
		switch word := headers[2]; true {
		case strings.HasSuffix(word, "_G"):
			sheet = "2G"
		case strings.HasSuffix(word, "_U"):
			sheet = "3G"
		default:
			log.Panicln("[ERROR]", "headers in csv not as expected")
		}
		int_libs.CSV_insert(data, xlsx_in, sheet, start_row)
		last_csv_row_data := data[len(data)-1]
		date = last_csv_row_data[0]
		time = last_csv_row_data[1]
	}

	date = strings.Replace(date, "-", "", -1)
	time = strings.Replace(time, ":", "", -1)

	date_time = date + "T" + time
	return
}

// wraper function
func parse_stat_files() {
	stat_files, _ := ioutil.ReadDir(stat_dir)

	if len(stat_files) != 0 {
		log.Println("[INFO]", len(stat_files), "files to proced in STAT dir")
		for _, file := range stat_files {
			if strings.HasSuffix(file.Name(), ".tgz") {
				log.Println("[INFO]", "extracting file=", stat_dir+file.Name())
				int_libs.Untgz(stat_dir+file.Name(), stat_dir, true)
			}
		}
		// section to call scripts
		var concurrency string
		concurrency = "parallel"
		switch {
		case concurrency == "none":
			log.Println("[INFO]", "calling NONE threads")
			for _, pr := range []string{"my.ref2.safmm.910.2Greport.counters", "my.ref2.safmm.910.3Greport.counters"} {
				call_script(pr)
			}

		case (concurrency == "thread") || (concurrency == "parallel"):
			switch concurrency {
			case "thread":
				log.Println("[INFO]", "runtime.GOMAXPROCS(1)")
				runtime.GOMAXPROCS(1)
			case "parallel":
				// at that moment use only 2 core (in this reason used keywoard "parallel" instead "2", "3", etc.)
				log.Println("[INFO]", "runtime.GOMAXPROCS(2)")
				runtime.GOMAXPROCS(2)
			}

			var wg sync.WaitGroup
			wg.Add(2)
			log.Println("[INFO]", "calling threads")
			go func() {
				defer wg.Done()
				defer log.Println("[DEBUG]", "finifhed thread #1")
				log.Println("[DEBUG]", "started thread #1")
				call_script("my.ref2.safmm.910.2Greport.counters")

			}()

			go func() {
				defer wg.Done()
				defer log.Println("[DEBUG]", "finifhed thread #2")
				log.Println("[DEBUG]", "started thread #2")
				call_script("my.ref2.safmm.910.3Greport.counters")
			}()

			wg.Wait()
			log.Println("[INFO]", "finished all threads")

			if ! *keepSTAT {
				stat_files, _ := ioutil.ReadDir(stat_dir)
				for _, file := range stat_files {
					file_name := stat_dir + file.Name()
					log.Println("[DEBUG]", "removing from STAT dir file=", file_name)
					os.Remove(file_name)
				}
				log.Println("[INFO]", "revoved ALL files from STAT dir")
			}

		}

	} else {
		log.Println("[WARN]", len(stat_files), "files to proced in STAT dir, skipping....")
	}
}

func check_CSV_dir() {
	csv_files, _ := ioutil.ReadDir(csv_dir)
	reports_number := len(csv_files)
	switch {
	case reports_number > 2:
		log.Println("[ERROR]", len(csv_files), "files to proced in SCV dir")
		log.Println("[ERROR]", "!!!!!!!!!!!! more then 2 files to proced in SCV dir !!!!!!!!!!!!!!!!")
		os.Exit(1)
	case reports_number == 0:
		log.Println("[ERROR]", "!!!!!!!!!!!! NOTHING to proced in SCV dir !!!!!!!!!!!!!!!!")
		os.Exit(1)
	case reports_number == 2:
		log.Println("[INFO]", reports_number, "files to proced in SCV dir")
	default:
		log.Println("ERROR", "wrong number of files in SCV dir, =", reports_number)
		os.Exit(1)
	}
}

func cleanCSV_dir () {
	csv_files, _ := ioutil.ReadDir(csv_dir)
	//	delete CSV csv_files fromm CSV dir
		if  ! *keepCSV {
			for _, file := range csv_files {
				file_name := csv_dir + file.Name()
				log.Println("[INFO]", "removing from CSV dir file=", file_name)
				os.Remove(file_name)
			}
		}


}

var help_message =
	"for work should be predifined folder structure\n" +
"   ./   \n" +
"    │   com_obj.exe\n" +
"    │\n" +
"    ├───CSV\n" +
"    ├───REPORTS\n" +
"    ├───SCRIPTS\n" +
"    │       my.ref2.safmm.910.2Greport.counters\n" +
"    │       my.ref2.safmm.910.3Greport.counters\n" +
"    │       SGSN_template.xlsx\n" +
"    │       wxp.pl\n" +
"    │\n" +
"    └───STAT\n"



var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file (for debug purpose only)")
var help = flag.Bool("help", false, "show long help info")
var keepCSV = flag.Bool("keep-csv", false, "don't delete CSV files after process")
var keepSTAT = flag.Bool("keep-stat", false, "don't remove files in STAT after process")
var version = flag.Bool("v", false, "show version and exit")
var csv_only = flag.Bool("csv-only", false, "proced only CSV files")
var debug_flag = flag.String("d", "INFO", "debug level, available levels: " +
	strings.Join(debug_levels, " "))

func main() {
	flag.Parse()

	// generate []logutils.LogLevel from []string
	debug_levels_f := make([]logutils.LogLevel, len(debug_levels))
	for _, level := range debug_levels {
		debug_levels_f = append(debug_levels_f, logutils.LogLevel(level))
	}

	filter := &logutils.LevelFilter{
		Levels: []logutils.LogLevel(debug_levels_f),
		MinLevel: logutils.LogLevel(*debug_flag),
		Writer: os.Stderr,
	}
	log.SetOutput(filter)

	site_name := set_site_name()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		check(err)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *help {
		fmt.Println(help_message)
		os.Exit(0)
	}

	if *version {
		fmt.Println(ver_string)
		os.Exit(0)
	}

	if *keepCSV {
		log.Println("[DEBUG]", "requested to keep CSV files")
	}

	if *csv_only {
		ReportCreator(site_name)
		cleanCSV_dir()
		os.Exit(0)
	}
	//------------------------------------------------
	parse_stat_files()
	check_CSV_dir()
	ReportCreator(site_name)
	cleanCSV_dir()
}
