// some internal functions
package int_libs

import (
	"os"
	"encoding/csv"
	"github.com/xuri/excelize"
	"strconv"
	"log"
	//"fmt"
	//"github.com/tealeg/xlsx"
)

// funtions to insert csv into xlsx template

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

func CSV_insert(data [][]string, xlsxFile *excelize.File, sheet string, start_row int) {
	//
	//_, data := read_csv_file(csv_file)

	//xlsxFile, err := excelize.OpenFile(xlsx_in_file)
	//check(err)
	//fmt.Printf(xlsxFile.GetCellFormula("2G", "AH2"))
	//cell2format := xlsxFile.GetCellStyle(sheet, "C2")
	//cell3format := xlsxFile.GetCellStyle(sheet, "AH2")
	style_number, err := xlsxFile.NewStyle(`{"number_format": 3}`)
	check(err)
	//fmt.Printf("cell format", cell2format, "---cell3", cell3format)
	for row_i, row_data := range data {
		for column, cell_data := range row_data {
			row := start_row + row_i
			cell := excelize.ToAlphaString(column) + strconv.Itoa(row)
			if column <= 1 {
				log.Println("[DEBUG]", "working on row=", row, "column=", column, "cell", cell)
				xlsxFile.SetCellValue(sheet, cell, cell_data)

			} else {
				log.Println("[DEBUG]", "working on row=", row, "column=", column, "cell", cell)
				f, err := strconv.ParseFloat(cell_data, 32)
				check(err)
				xlsxFile.SetCellValue(sheet, cell, f)
				xlsxFile.SetCellStyle(sheet, cell, cell, style_number)
			}
			if sheet == "2G" {
			xlsxFile.SetCellFormula(sheet, "AH" + strconv.Itoa(row), Formulas2G(row, "AH"))
			xlsxFile.SetCellFormula(sheet, "AI" + strconv.Itoa(row), Formulas2G(row, "AI"))
			xlsxFile.SetCellFormula(sheet, "AJ" + strconv.Itoa(row), Formulas2G(row, "AJ"))
			xlsxFile.SetCellFormula(sheet, "AK" + strconv.Itoa(row), Formulas2G(row, "AK"))
			xlsxFile.SetCellFormula(sheet, "AL" + strconv.Itoa(row), Formulas2G(row, "AL"))
			}

			if sheet == "3G" {
				xlsxFile.SetCellFormula(sheet, "AM" + strconv.Itoa(row), Formulas3G(row, "AM"))
				xlsxFile.SetCellFormula(sheet, "AN" + strconv.Itoa(row), Formulas3G(row, "AN"))
				xlsxFile.SetCellFormula(sheet, "AO" + strconv.Itoa(row), Formulas3G(row, "AO"))
				xlsxFile.SetCellFormula(sheet, "AP" + strconv.Itoa(row), Formulas3G(row, "AP"))
				xlsxFile.SetCellFormula(sheet, "AQ" + strconv.Itoa(row), Formulas3G(row, "AQ"))
				xlsxFile.SetCellFormula(sheet, "AR" + strconv.Itoa(row), Formulas3G(row, "AR"))
			}

			//fmt.Printf("cell", cell, "formula", xlsxFile.GetCellFormula("2G", "AH2"))
		}
	}
}



