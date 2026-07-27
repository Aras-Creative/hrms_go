package usecase

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"hrms/internal/pkg/timeutil"
	errors "hrms/internal/pkg/apperror"
)

func (uc *DailyAttendanceUsecase) GenerateReportXLSX(ctx context.Context, fromStr, toStr string) ([]byte, string, error) {
	report, err := uc.GetReportData(ctx, fromStr, toStr)
	if err != nil {
		return nil, "", err
	}

	loc := timeutil.LoadDefaultLocation()
	from := report.From.In(loc)
	daysInMonth := time.Date(from.Year(), from.Month()+1, 0, 0, 0, 0, 0, loc).Day()

	monthName := indonesianMonths[from.Month()]
	title := fmt.Sprintf("REKAP KEHADIRAN CV ARAS CREATIVE BULAN %s %d", monthName, from.Year())
	fullWidth := 5 + daysInMonth*3

	f := excelize.NewFile()
	f.SetSheetName("Sheet1", "Rekap Kehadiran")
	sheet := "Rekap Kehadiran"

	// Freeze panes (row 1-3 frozen, col A frozen)
	f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		XSplit:      1,
		YSplit:      3,
		TopLeftCell: "B4",
	})

	thinBorder := []excelize.Border{
		{Type: "top", Style: 1, Color: "000000"},
		{Type: "bottom", Style: 1, Color: "000000"},
		{Type: "left", Style: 1, Color: "000000"},
		{Type: "right", Style: 1, Color: "000000"},
	}
	centerAlign := &excelize.Alignment{Horizontal: "center", Vertical: "center"}
	leftAlign := &excelize.Alignment{Horizontal: "left", Vertical: "center"}

	makeStyle := func(font *excelize.Font, fill excelize.Fill, align *excelize.Alignment) int {
		id, _ := f.NewStyle(&excelize.Style{
			Font: font, Fill: fill, Alignment: align, Border: thinBorder,
		})
		return id
	}
	setCell := func(col, row int, val string) string {
		ref := colName(col) + fmt.Sprint(row)
		f.SetCellValue(sheet, ref, val)
		return ref
	}
	setStyle := func(ref string, styleID int) {
		_ = f.SetCellStyle(sheet, ref, ref, styleID)
	}

	// Row 1: Title
	titleStyle := makeStyle(
		&excelize.Font{Bold: true, Size: 16},
		excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"B4C6E7"}},
		centerAlign,
	)
	titleEnd := colName(fullWidth) + "1"
	f.MergeCell(sheet, "A1", titleEnd)
	f.SetCellValue(sheet, "A1", title)
	f.SetCellStyle(sheet, "A1", titleEnd, titleStyle)

	// Row 2: Main header
	yellowFill := excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"FFE699"}}
	blueFill := excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"BDD7EE"}}
	greenFill := excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"E2EFDA"}}
	hdrNameStyle := makeStyle(&excelize.Font{Bold: true}, yellowFill, centerAlign)
	hdrDateStyle := makeStyle(&excelize.Font{Bold: true}, blueFill, centerAlign)
	hdrSubStyle := makeStyle(&excelize.Font{Bold: true}, greenFill, centerAlign)

	col := 1
	for _, v := range []string{"Nama Karyawan", "Hadir", "Sakit", "Izin/Cuti", "Absen"} {
		ref := setCell(col, 2, v)
		if col == 1 {
			setStyle(ref, hdrNameStyle)
		} else {
			setStyle(ref, hdrDateStyle)
		}
		col++
	}
	for d := 1; d <= daysInMonth; d++ {
		startCol := col
		for _, v := range []string{fmt.Sprintf("%d %s", d, monthName), "", ""} {
			ref := setCell(col, 2, v)
			setStyle(ref, hdrDateStyle)
			col++
		}
		f.MergeCell(sheet, colName(startCol)+"2", colName(startCol+2)+"2")
	}

	// Row 3: Sub header
	col = 1
	for range 5 {
		ref := setCell(col, 3, "")
		if col == 1 {
			setStyle(ref, hdrNameStyle)
		} else {
			setStyle(ref, hdrSubStyle)
		}
		col++
	}
	for d := 1; d <= daysInMonth; d++ {
		for _, v := range []string{"Masuk", "Pulang", "Status"} {
			ref := setCell(col, 3, v)
			setStyle(ref, hdrSubStyle)
			col++
		}
	}

	// Merge header cells vertically (row 2-3)
	for _, c := range []string{"A", "B", "C", "D", "E"} {
		f.MergeCell(sheet, c+"2", c+"3")
	}

	// Auto filter on row 3
	f.AutoFilter(sheet, "A3:"+colName(fullWidth)+"3", nil)

	// Summary column styles
	summaryStyles := map[int]int{
		2: makeStyle(&excelize.Font{Bold: true, Color: "006100"}, excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"CCFFCC"}}, centerAlign),
		3: makeStyle(&excelize.Font{Bold: true, Color: "9C6500"}, excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"FFFF99"}}, centerAlign),
		4: makeStyle(&excelize.Font{Bold: true, Color: "2F75B5"}, excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"DDEBF7"}}, centerAlign),
		5: makeStyle(&excelize.Font{Bold: true, Color: "9C0006"}, excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"FFCCCC"}}, centerAlign),
	}

	type sColors struct{ fill, font string; bold bool }
	statusColorMap := map[string]sColors{
		"sakit":       {"FFFF99", "9C6500", true},
		"hadir":       {"CCFFCC", "006100", true},
		"terlambat":   {"FFE5CC", "9C6500", true},
		"pulang awal": {"D3D3D3", "000000", false},
		"libur":       {"D3D3D3", "000000", false},
	}
	defColors := sColors{"DDEBF7", "2F75B5", true}
	absColors := sColors{"FFCCCC", "9C0006", true}

	dataRowStart := 4
	for rowIdx, emp := range report.Employees {
		row := dataRowStart + rowIdx
		totalHadir, totalSakit, totalIzin, totalAbsen := 0, 0, 0, 0

		logMap := make(map[int]ReportDailyLog)
		for _, log := range emp.DailyLogs {
			logMap[log.Date.In(loc).Day()] = log
		}

		dailyData := make([]string, 0, daysInMonth*3)
		for d := 1; d <= daysInMonth; d++ {
			clockIn, clockOut, status := "-", "-", "-"
			if log, ok := logMap[d]; ok {
				clockIn, clockOut, status = log.ClockIn, log.ClockOut, log.Status
			}

			sl := strings.ToLower(status)
			switch {
			case sl == "sakit":
				totalSakit++
			case sl == "hadir" || sl == "terlambat" || sl == "pulang awal" || strings.Contains(sl, "berangkat siang"):
				totalHadir++
			case sl == "" || sl == "-" || sl == "tidak masuk":
				totalAbsen++
			case sl == "libur" || sl == "day_off":
			default:
				totalIzin++
			}
			dailyData = append(dailyData, clockIn, clockOut, status)
		}

		rowValues := append([]interface{}{emp.EmployeeName, totalHadir, totalSakit, totalIzin, totalAbsen},
			toInterface(dailyData)...)

		for colIdx, val := range rowValues {
			f.SetCellValue(sheet, colName(colIdx+1)+fmt.Sprint(row), val)
		}

		for colIdx := 1; colIdx <= fullWidth; colIdx++ {
			cell := colName(colIdx) + fmt.Sprint(row)
			switch {
			case colIdx == 1:
				setStyle(cell, makeStyle(nil, excelize.Fill{}, leftAlign))
			case colIdx >= 2 && colIdx <= 5:
				if s, ok := summaryStyles[colIdx]; ok {
					setStyle(cell, s)
				}
			case colIdx > 5 && (colIdx-2)%3 == 0:
				valStr := fmt.Sprintf("%v", rowValues[colIdx-1])
				sl := strings.ToLower(valStr)

				var sc sColors
				switch {
				case sl == "" || sl == "-" || sl == "tidak masuk":
					sc = absColors
				case strings.Contains(sl, "sakit"):
					sc = statusColorMap["sakit"]
				case sl == "hadir":
					sc = statusColorMap["hadir"]
				case strings.Contains(sl, "terlambat"):
					sc = statusColorMap["terlambat"]
				case strings.Contains(sl, "pulang"):
					sc = statusColorMap["pulang awal"]
				case sl == "libur":
					sc = statusColorMap["libur"]
				default:
					sc = defColors
				}
				setStyle(cell, makeStyle(
					&excelize.Font{Bold: sc.bold, Color: sc.font},
					excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{sc.fill}},
					centerAlign,
				))
			default:
				setStyle(cell, makeStyle(nil, excelize.Fill{}, centerAlign))
			}
		}
	}

	// Column widths
	maxNameLen := 10
	for _, emp := range report.Employees {
		if l := len(emp.EmployeeName); l > maxNameLen {
			maxNameLen = l
		}
	}
	f.SetColWidth(sheet, "A", "A", float64(maxNameLen)+1.5)
	for i := 2; i <= fullWidth; i++ {
		f.SetColWidth(sheet, colName(i), colName(i), 9)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, "", errors.NewInternal("failed to write excel: " + err.Error())
	}

	fileName := fmt.Sprintf("attendance_recap_%s_%d.xlsx", strings.ToLower(monthName), from.Year())
	return buf.Bytes(), fileName, nil
}

var indonesianMonths = map[time.Month]string{
	time.January: "Januari", time.February: "Februari", time.March: "Maret",
	time.April: "April", time.May: "Mei", time.June: "Juni",
	time.July: "Juli", time.August: "Agustus", time.September: "September",
	time.October: "Oktober", time.November: "November", time.December: "Desember",
}

func colName(n int) string {
	result := ""
	for n > 0 {
		n--
		result = string(rune('A'+n%26)) + result
		n /= 26
	}
	return result
}

func toInterface(s []string) []interface{} {
	r := make([]interface{}, len(s))
	for i, v := range s {
		r[i] = v
	}
	return r
}
