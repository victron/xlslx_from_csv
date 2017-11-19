package main

import (
	"flag"
	"strings"
	"os"
	"log"
	"runtime/pprof"
	"fmt"
)

var help_message = "for work should be predifined folder structure\n" +
	"   ./   \n" +
	"    │   stat_wraper.exe\n" +
	"    │\n" +
	"    ├───CSV\n" +
	"    ├───REPORTS\n" +
	"    ├───SCRIPTS\n" +
	"    │       my.ref2.safmm.910.2Greport.counters\n" +
	"    │       my.ref2.safmm.910.3Greport.counters\n" +
	"    │       SGSN_template.xlsx\n" +
	"    │       wxp.pl\n" +
	"    │\n" +
	"    └───STAT\n" +
	"\n" +
	" to work in auto mode use auto subcommand, use \"auto -h\" to get help on flags\n" +
	"example: stat_wraper.exe auto -h 10.156.48.117 -p 22 -u vtsymbal -s sexyterm1 -d DEBUG"

var debug = "INFO"
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file (for debug purpose only)")
var help = flag.Bool("help", false, "show long help info")
var keepCSV = flag.Bool("keep-csv", false, "don't delete CSV files after process")
var keepSTAT = flag.Bool("keep-stat", false, "don't remove files in STAT after process")
var version = flag.Bool("v", false, "show version and exit")
var csv_only = flag.Bool("csv-only", false, "proced only CSV files")

// automatic method initialize
var auto_task = false
//var auto = flag.Bool("auto", false, "connect to server and auto download files")
var auto = flag.NewFlagSet("auto", flag.ExitOnError)
var host = auto.String("h", "", "hostname or IP addres of server (mandatory)")
var port = auto.Int("p", 22, "server port")
var user = auto.String("u", "", "login to connect to server (mandatory)")
var pass = auto.String("s", "", "sercret (password) to connect (mandatory)")
var interval = auto.Int("i", 15, "how often to build report, allowed values [15, 30, 60]")

func init() {
	switch os.Args[1] {
	case "auto":
		// debug flag for "auto" subcommand
		auto.StringVar(&debug, "d", "INFO", "debug level, available levels: "+strings.Join(debug_levels, " "))
		// parsing
		auto.Parse(os.Args[2:])
		if *host == "" || *user == "" || *pass == "" {
			auto_task = false
			log.Fatal("[FATAL]", " please check mandatory flags using auto subcommand")
		}
		server.Ip = *host
		server.Port = *port
		server.UserName = *user
		server.UserPass = *pass

		auto_task = true
		default:
			// debug flag without subcommand
			flag.StringVar(&debug, "d", "INFO", "debug level, available levels: "+strings.Join(debug_levels, " "))
			flag.Parse()

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
		}
	}
