//go:generate go run generate_embed_list.go
//go:generate gofmt -w embedded_files.go

package main

import (
	"bytes"
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/pkg/errors"

	"github.com/xuri/excelize/v2"
)

var (
	Version  string = "0.0.1"
	Revision        = func() string { // {{{
		revision := ""
		modified := false
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" {
					//return setting.Value
					revision = setting.Value
					if len(setting.Value) > 7 {
						revision = setting.Value[:7] // 最初の7文字にする
					}
				}
				if setting.Key == "vcs.modified" {
					modified = setting.Value == "true"
				}
			}
		}
		if modified {
			revision = "develop+" + revision
		}
		return revision
	}() // }}}

	embeddedFiles *embed.FS   // go:generateで生成されるembedded_fils.go内のinit()で代入される。
	parser        *arg.Parser // ShowHelp() で使う
	debug_mode    = false     // trueの時ログ出力を出す。
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	args := argparse()

	log.Printf("args: %v", args)

	if len(args.Input) == 0 {
		//FromStdin(args) // 標準入力から読み取り、標準出力に出力する。
		// xlsxをcatして与えられることは想定しない。
		return
	} else {
		for _, path := range args.Input { // 引数で与えられたファイルを1つづつ処理する。
			// defer func() {
			// 	if rec := recover(); rec != nil {
			// 		log.Printf("path: %v", path)
			// 		log.Printf("Recovered from: %w", rec)
			// 	}
			// }()

			// 拡張子チェックしない。
			//if ext := strings.ToLower(filepath.Ext(path)); ext != ".xlsx" {
			//	log.Printf("path: %#v", path)
			//	continue
			//}
			ExcelToCsv(path, args)
		}
	}
}

func ShowHelp(post string) {
	buf := new(bytes.Buffer)
	parser.WriteHelp(buf)
	fmt.Printf("%v\n", strings.ReplaceAll(buf.String(), "display this help and exit", "ヘルプを出力する。"))
	if len(post) != 0 {
		fmt.Println(post)
	}
	os.Exit(1)
}
func ShowVersion() {
	if len(Revision) == 0 {
		// go installでビルドされた場合、gitの情報がなくなる。その場合v0.0.0.のように末尾に.がついてしまうのを避ける。
		fmt.Printf("%v version %v\n", GetFileNameWithoutExt(os.Args[0]), Version)
	} else {
		fmt.Printf("%v version %v.%v\n", GetFileNameWithoutExt(os.Args[0]), Version, Revision)
	}
	os.Exit(0)
}

func ExcelToCsv(path string, args *Args) {

	log.Printf("path: %#v", path)

	f, err := excelize.OpenFile(path)
	if err != nil {
		//panic(errors.Errorf("%v", err))
		log.Printf("%v", errors.Errorf("%v", err))
		return
	}

	// 引数で、シート番号が指定できる。指定されているとき、指定されたシートのみ変換する。
	// mapのキーに指定されたシート番号を入れてfor文内で指定されているか判定する。
	m := map[int]int{}
	for _, s := range args.Sheet {
		m[s] = s
	}
	log.Printf("%v", m)

	sheetList := f.GetSheetList()
	for i := 0; i < len(sheetList); i++ {
		if len(args.Sheet) != 0 { // 変換するシート番号が引数で指定されているとき、指定されていないシート番号はスキップする。

			exist := func() bool {
				for _, sheet_num := range args.Sheet {
					if i == sheet_num {
						return true
					}
				}
				return false
			}()

			if !exist {
				continue
			}
			log.Printf("sheet_num: %v", sheetList[i])
		}
		v := sheetList[i]
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

		log.Printf("%v", filename)
	}
}

func GetFileNameWithoutExt(path string) string {
	return filepath.Base(path[:len(path)-len(filepath.Ext(path))])
}
func GetFilePathWithoutExt(path string) string {
	return filepath.ToSlash(filepath.Join(filepath.Dir(path), GetFileNameWithoutExt(path)))
}

func argparse() *Args {
	args := &Args{}

	var err error
	parser, err = arg.NewParser(arg.Config{Program: GetFileNameWithoutExt(os.Args[0]), IgnoreEnv: false}, args)
	if err != nil {
		ShowHelp(fmt.Sprintf("%v", errors.Errorf("%v", err)))
	}
	if err := parser.Parse(os.Args[1:]); err != nil {
		if err.Error() == "help requested by user" {
			ShowHelp("")
		} else if err.Error() == "version requested by user" {
			ShowVersion()
		} else {
			panic(errors.Errorf("%v", err))
		}
	}
	if args.Version {
		ShowVersion()
	}
	if args.Code {
		WriteEmbeddedData("./code")
		os.Exit(0)
	}
	if args.Debug {
		debug_mode = true // trueの時ログ出力を出す。
	}
	if !debug_mode {
		log.SetOutput(io.Discard)
	}
	return args
}

func WriteEmbeddedData(outputPath string) {
	err := fs.WalkDir(*embeddedFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.Errorf("failed to access %s: %v", path, err)
		}
		if d.IsDir() {
			return nil
		}
		data, err := embeddedFiles.ReadFile(path)
		if err != nil {
			return errors.Errorf("failed to read embedded file %s: %v", path, err)
		}
		outPath := filepath.ToSlash(filepath.Join(outputPath, path))
		if err := os.MkdirAll(filepath.Dir(outPath), os.ModePerm); err != nil {
			return errors.Errorf("failed to create directories for %s: %v", outPath, err)
		}
		if err := os.WriteFile(outPath, data, 0644); err != nil {
			return errors.Errorf("failed to write file %s: %v", outPath, err)
		}
		log.Printf("%v", outPath)
		return nil
	})
	if err != nil {
		panic(errors.Errorf("failed to walk embedded files: %v", err))
	}
}

func (args Args) String() string {
	return fmt.Sprintf(`
	Input   : %#v
	Debug   : %v
	Version : %v
	Sheet   : %v
	`,
		args.Input,
		args.Debug,
		args.Version,
		args.Sheet,
	)
}

type Args struct {
	Input   []string `arg:"positional" help:"csvに変換するxlsxを指定する。"`
	Sheet   []int    `arg:"-s,--sheet,separate" help:"csvに変換するxlsxのシートを番号で指定する。指定がないときすべてのシートを変換する。"`
	Code    bool     `arg:"-c,--code"     help:"このプログラムのソースコードを出力する。./codeディレクトリに出力される。"`
	Debug   bool     `arg:"-d,--debug" help:"デバッグ用。ログが詳細になる。"`
	Version bool     `arg:"-v,--version" help:"バージョン情報を出力する。"`
}
