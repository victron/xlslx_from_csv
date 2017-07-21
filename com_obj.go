package main

import (
	"fmt"
	"os"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
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
)

var ver_string = "version = 0.2.2, build date = 2017-07-21"

// TODO: flags for different concurrency modes

const xlWorkbookDefault = 51
var cwd, _ = os.Getwd()
var reports_dir = cwd + "\\REPORTS\\"
var scripts_dir  = cwd + "\\SCRIPTS\\"
var csv_dir = cwd + "\\CSV\\"
var stat_dir = cwd + "\\STAT\\"

var sheets_ids = map[int]string{1: "graph", 2: "2G", 3: "3G"}
var sheets_names = map[string]int{"graph": 1, "2G": 2, "3G": 3}

func check(e error) {
	if e != nil {
		log.Panicln("[DEBUG]", "error=", e)
	}
}

// return: int: last row on worksheet
func last_row(worksheet *ole.IDispatch) int {
	used_range := oleutil.MustGetProperty(worksheet, "UsedRange" ).ToIDispatch()
	rows := oleutil.MustGetProperty(used_range, "Rows" ).ToIDispatch()
	rowscount := (int)(oleutil.MustGetProperty(rows, "Count").Val)
	log.Println("[DEBUG] rows count=", rowscount)
	used_range.Release()
	rows.Release()
	return rowscount
}

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

// checks template or exist document
// return: int - number of rows in template, map[string]int - sheet index for 2G or 3G
func check_xlsx_file(fileName string, excel, workbooks *ole.IDispatch) bool {

	workbook, err := oleutil.CallMethod(workbooks, "Open", fileName)
	check(err)
	defer workbook.ToIDispatch().Release()

	sheets := oleutil.MustGetProperty(excel, "Sheets").ToIDispatch()
	sheetCount := (int)(oleutil.MustGetProperty(sheets, "Count").Val)
	log.Println("[DEBUG] sheet count=", sheetCount)
	sheets.Release()

	if sheetCount != 3 {
		log.Println("[CRIT]", "wrong number of sheets")
		return false
	}
	for sheet_num, sheet_name := range sheets_ids {
		worksheet := oleutil.MustGetProperty(workbook.ToIDispatch(), "Worksheets", sheet_num).ToIDispatch()
		val, err := oleutil.GetProperty(worksheet, "Name")
		log.Println("[DEBUG]", "worksheet name=", val.Value(), "sheet_num=", sheet_num)
		defer worksheet.Release()
		check(err)
		if sheet_name != val.Value() {
			return false
		}

	}
return true
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

// input: fileName - file with prefious stat,
func read_write(fileName string, excel, workbooks *ole.IDispatch,  ) {
	//const xlExcel8 = 56

	if ! check_xlsx_file(reports_dir + fileName, excel, workbooks) {	// TODO: move into main
		log.Panicln("[ERROR]", "xlsx file NOK, file=", reports_dir + fileName)
	}

	workbook, err := oleutil.CallMethod(workbooks, "Open", reports_dir + fileName)
	check(err)
	defer workbook.ToIDispatch().Release()


	//var csv_file string

	file_s_csv, _ := ioutil.ReadDir("CSV")
	var last_row_data []string

	// prepare filename prefix
	fileName_no_extention := strings.Split(fileName, ".")[0]
	file_name_prefixs := strings.Split(fileName_no_extention, "_")
	file_name_prefix := strings.Join(file_name_prefixs[:2], "_")

	for _, csv_file := range file_s_csv {
		headers, data := read_csv_file(csv_dir + csv_file.Name())
		var sheet_name string

		switch word := headers[2]; true {
		case strings.HasSuffix(word, "_G"):
			sheet_name = "2G"
		case strings.HasSuffix(word, "_U"):
			sheet_name = "3G"
		default:
			log.Panicln("[ERROR]", "headers in csv not as expected")
		}

		// TODO: move to last_row func and compare numbers on both sheets

		worksheet := oleutil.MustGetProperty(workbook.ToIDispatch(), "Worksheets",
			sheets_names[sheet_name]).ToIDispatch()
		defer worksheet.Release()

		//var last_row int
		last_row := last_row(worksheet)

		//row := 112
		//col := 2
		//cell := oleutil.MustGetProperty(worksheet, "Cells", row, col).ToIDispatch()
		//val, err := oleutil.GetProperty(cell, "Value")
		//fmt.Printf("(%d,%d)=%+v toString=%s\n", col, row, val.Value(), val.Val)
		//cell.Release()

		for row_i, row_data := range data {
			for column, cell_data := range row_data {
				//log.Println("[DEBUG]", "working on row=", row_i + last_row + 1, "column=", column + 1)
				var row int
				if len(file_name_prefixs) == 2 {
					row = row_i+last_row
				} else {
					row = row_i+last_row + 1
				}
				cell := oleutil.MustGetProperty(worksheet, "Cells", row, column+1).ToIDispatch()
				oleutil.PutProperty(cell, "Value", cell_data)
				cell.Release()
			}

		}

		last_row_data = data[len(data)-1]
	}

	date := last_row_data[0]
	time := last_row_data[1]
	date = strings.Replace(date, "-", "", -1)
	time = strings.Replace(time, ":", "", -1)

	new_file_name := cwd + "\\REPORTS\\" + file_name_prefix + "_" + date + "_" + time + ".xlsx"

	activeWorkBook := oleutil.MustGetProperty(excel, "ActiveWorkBook").ToIDispatch()
	defer activeWorkBook.Release()
	const xlExcel8 = 56
	const xlExcel12 = 50
	const xlWorkbookDefault = 51
	log.Println("[INFO]", "SavaAs=", new_file_name)
	oleutil.MustCallMethod(activeWorkBook, "SaveAs", new_file_name, xlWorkbookDefault, nil, nil).ToIDispatch()
	defer oleutil.MustCallMethod(activeWorkBook, "Close").ToIDispatch()
	//workbook := oleutil.MustCallMethod(workbooks, "Add", nil).ToIDispatch()
}

// function to copy file
func copy_file(src string, dst string) {
	// Read all content of src to data
	data, err := ioutil.ReadFile(src)
	check(err)
	// Write data to dst
	err = ioutil.WriteFile(dst, data, 0644)
	check(err)
}

// check dir REPORTS, if it's epmty use SGSN_template.xlsx to create new report file
// return: string - working file KPI-SGSN_<SITE_NAME>_YYYYmmdd_hhmm.xlsx
//
func working_file( excel, workbooks *ole.IDispatch) (last_report_file string) {
	files, _ := ioutil.ReadDir("REPORTS")
	var file_name_prefix string
	//var last_csv_file string
	//var last_report_file string
	if len(files) == 0 {
		fmt.Println("previous reports not found in \"REPORTS\" dir. \n" +
			"Need to know SiteName. (for example: KIE1)")
		fmt.Print("SiteName: ")
		var site_name string
		fmt.Scanf("%s", &site_name)
		file_name_prefix = "KPI-SGSN_" + site_name
		log.Println("[DEBUG]", "opening file=", scripts_dir + "SGSN_template.xlsx")
		workbook, err := oleutil.CallMethod(workbooks, "Open", scripts_dir + "SGSN_template.xlsx")
		check(err)
		defer workbook.ToIDispatch().Release()

		worksheet := oleutil.MustGetProperty(workbook.ToIDispatch(), "Worksheets", sheets_names["graph"]).ToIDispatch()
		defer worksheet.Release()

		cell := oleutil.MustGetProperty(worksheet, "Cells", 1, 10).ToIDispatch()
		oleutil.PutProperty(cell, "Value", site_name)
		cell.Release()

		last_report_file = file_name_prefix +".xlsx"
		activeWorkBook := oleutil.MustGetProperty(excel, "ActiveWorkBook").ToIDispatch()
		log.Println("[INFO]", "SaveAs=", reports_dir + last_report_file)
		oleutil.MustCallMethod(activeWorkBook, "SaveAs", reports_dir + last_report_file, xlWorkbookDefault, nil, nil).ToIDispatch()
		activeWorkBook.Release()
		oleutil.MustCallMethod(activeWorkBook, "Close").ToIDispatch()

	} else {
		// DONE: ~$KPI-SGSN_KIE#1_20170608_0600.xlsx ignore such files
		var allowed_files  []os.FileInfo
		for _, file := range files{
			if  strings.HasPrefix(file.Name(), file_name_prefix){
				allowed_files = append(allowed_files, file)
			} else {
				log.Println("[WARN]", "found not allowed filename=", file.Name(), "in REPORTS")
			}
		}
		last_report_file = allowed_files[len(files) -1 ].Name()
		log.Println("[INFO]", "last report file with index = -1", last_report_file)

	}
	return
}

// wraper function
func write_xlsx(excel, workbooks *ole.IDispatch) {
	last_report_file := working_file(excel, workbooks) // working_file checks if reports file exists, so better call here

	stat_files, _ := ioutil.ReadDir(stat_dir)

	if len(stat_files) != 0 {
		log.Println("[INFO]", len(stat_files), "files to proced in STAT dir")
		for _, file := range stat_files {
			if strings.HasSuffix(file.Name(), ".tgz") {
				log.Println("[INFO]", "extracting file=", stat_dir + file.Name())
				targz.Untgz(stat_dir + file.Name(), stat_dir, true)
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

	csv_files, _ := ioutil.ReadDir(csv_dir)

	if len(csv_files) != 0 {
		log.Println("[INFO]", len(csv_files), "files to proced in SCV dir")
		if len(csv_files) > 2 {
			log.Println("[WARN]", len(csv_files), "files to proced in SCV dir")
			log.Println("[WARN]", "!!!!!!!!!!!! more then 2 files to proced in SCV dir !!!!!!!!!!!!!!!!")
		}

		read_write(last_report_file, excel, workbooks)
	//	delete CSV csv_files fromm CSV dir
		if  ! *keepCSV {
			for _, file := range csv_files {
				file_name := csv_dir + file.Name()
				log.Println("[INFO]", "removing from CSV dir file=", file_name)
				os.Remove(file_name)
			}
		}

	} else {
		log.Println("[ERR]", len(csv_files), "files to proced in SCV dir")
		log.Println("[ERR]", "!!!!!!!!!!!! NOTHING to proced in SCV dir !!!!!!!!!!!!!!!!")
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

func main() {
	//debug_levels := []logutils.LogLevel{"DEBUG", "WARN", "ERROR", "INFO"}
	debug_levels := []string{"DEBUG", "WARN", "ERROR", "INFO"}
	var debug_flag = flag.String("d", "INFO", "debug level, available levels: " +
												strings.Join(debug_levels, " "))
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

	//------------------------------------------------
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, _ := oleutil.CreateObject("Excel.Application")
	excel, _ := unknown.QueryInterface(ole.IID_IDispatch)
	defer excel.Release()

	oleutil.PutProperty(excel, "Visible", false)
	defer oleutil.CallMethod(excel, "Quit")

	workbooks := oleutil.MustGetProperty(excel, "Workbooks").ToIDispatch()
	defer workbooks.Release()

	write_xlsx(excel, workbooks)
	//read_write(cwd+"\\KPI_SGSN_ZAP_20170523_1000.xlsx", excel, workbooks, cwd + "\\CSV" + "\\Service=safmm_34.csv")
	//showMethodsAndProperties(workbooks)
}
