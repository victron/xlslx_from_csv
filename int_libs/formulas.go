package int_libs

import (
	"strconv"
	"log"
)

func Formulas2G(row_int int, column string) string {

	row := strconv.Itoa(row_int)
	formula := ""
	switch column {
	case "AH":
		//	=(F2+J2+K2+L2)/E2
		formula = "=(F" + row + "+J" + row + "+K" + row + "+L" + row + ")/E" + row

	case "AI":
		//	=(F2+I2+J2+K2+L2)/E2
		formula = "=(F" + row + "+I" + row + "+J" + row + "+K" + row + "+L" + row + ")/E" + row

	case "AJ":
		//	=(V2+W2+X2+Y2)/U2
		formula = "=(V" + row + "+W" + row + "+X" + row + "+Y" + row + ")/U" + row

	case "AK":
		//=AB2*8/1024/900
		formula = "=AB" + row + "*8/1024/900"

	case "AL":
		//=AB2*8/1024/900
		formula = "=AB" + row + "*8/1024/900"
	}
	
	log.Println("[DEBUG]", "result 2G formula=", formula)
	return formula
}


func Formulas3G(row_int int, column string) string {

	row := strconv.Itoa(row_int)
	formula := ""
	switch column {
	case "AM":
		//	=(F2+K2 + I2 + L2 + Q2+ AI2 + AJ2)/(F2+K2 + I2 + L2 + Q2 + G2 + H2 + J2 + AK2)%
		formula = "=(F" + row + "+K" + row + "+I" + row + "+L" + row + "+Q" + row + "+AI" + row + "+AJ" + row +
			")/(F" + row + "+K" + row + "+I" + row + "+L" + row + "+Q" + row + "+G" + row + "+H" + row + "+J" + row +
				"+AK" + row + ")%"

	case "AN":
		//=(K2+F2)/E2%
		formula = "=(K" + row + "+F" + row + ")/E" + row + "%"

	case "AO":
		//=(W2+AC2+AB2)/(AC2+AB2+W2+Y2+X2+Z2+AA2+AD2+AE2+AF2)%
		formula = "=(W" + row + "+AC" + row + "+AB" + row + ")/(AC" + row + "+AB" + row + "+W" + row + "+Y" + row +
			"+X" + row + "+Z" + row + "+AA" + row + "+AD" + row + "+AE" + row + "+AF" + row + ")%"

	case "AP":
		//=ROUND(AG2*8/1024/900,2)
		formula = "=ROUND(AG" + row + "*8/1024/900,2)"

	case "AQ":
		//=ROUND(AH2*8/1024/900,2)
		formula = "=ROUND(AH" + row + "*8/1024/900,2)"

	case "AR":
		//=AP2+AQ2
		formula = "=AP" + row + "+AQ" + row
	}
	log.Println("[DEBUG]", "result 3G formula=", formula)
	return formula
}