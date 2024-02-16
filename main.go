package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/xuri/excelize/v2"
)

func ExcelToCsv(path string) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		//panic(errors.Errorf("%v", err))
		log.Printf("%v", errors.Errorf("%v", err))
		return
	}

	sheetList := f.GetSheetList()
	for i, v := range sheetList {
		filename := fmt.Sprintf("%v_%v_%v.csv", GetFilePathWithoutExt(path), i, v)
		file, err := os.Create(filename)
		if err != nil {
			//panic(errors.Errorf("%v", err))
			log.Printf("%v", errors.Errorf("%v", err))
			continue
		}
		defer file.Close() // 関数終了時にファイルを閉じる
		rows, err := f.Rows(v)
		if err != nil {
			//panic(errors.Errorf("%v", err))
			log.Printf("%v", errors.Errorf("%v", err))
			continue
		}
		csvw := csv.NewWriter(file)
		defer csvw.Flush()
		for rows.Next() {
			cols, err := rows.Columns()
			if err != nil {
				//panic(errors.Errorf("%v", err))
				log.Printf("%v", errors.Errorf("%v", err))
				continue
			}
			if err := csvw.Write(cols); err != nil {
				//panic(errors.Errorf("%v", err))
				log.Printf("%v", errors.Errorf("%v", err))
				continue
			}
		}
	}

}

func GetFileNameWithoutExt(path string) string {
	return filepath.Join(filepath.Dir(path), filepath.Base(path[:len(path)-len(filepath.Ext(path))]))
}
func GetFilePathWithoutExt(path string) string {
	return filepath.ToSlash(filepath.Join(filepath.Dir(path), GetFileNameWithoutExt(path)))
}

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile) // ログの出力書式を設定する
	for i, f := range os.Args {
		if i == 0 {
			continue
		}
		if ext := strings.ToLower(filepath.Ext(f)); ext != ".xlsx" {
			continue
		}
		ExcelToCsv(f)
	}
}
