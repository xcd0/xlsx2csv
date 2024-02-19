package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
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
	Revision        = func() string {
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
	}()
	args *Args
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile) // ログの出力書式を設定する
	ArgParse()
	for _, f := range args.Input {
		if ext := strings.ToLower(filepath.Ext(f)); ext != ".xlsx" {
			continue
		}
		ExcelToCsv(f)
	}
}

type Args struct {
	Input   []string `arg:"positional"         help:"csvに変換するxlsxを指定する。" default:{}`
	Debug   bool     `arg:"-d,--debug"         help:"デバッグ用。ログが詳細になる。"`
	Version bool     `arg:"-v,--version"       help:"バージョン情報を出力する。"`
	//VersionSub *ArgsVersion `arg:"subcommand:version" help:"バージョン情報を出力する。"`
	// 現状positionalな引数とサブコマンドは同居できないらしい。
}

func (args *Args) Print() {
	log.Printf(`
	Input   : %v
	Debug   : %v
	Version : %v
	`,
		args.Input,
		args.Debug,
		args.Version,
	)
}

type ArgsVersion struct {
}

func ArgParse() {
	args = &Args{
		Input:   []string{},
		Debug:   false,
		Version: false,
		//VersionSub: nil,
	}
	//var parser *arg.Parser
	//var err error
	parser, err := arg.NewParser(arg.Config{Program: GetFileNameWithoutExt(os.Args[0]), IgnoreEnv: false}, args)
	if err != nil {
		log.Printf("%v", errors.Errorf("%v", err))
		ShowHelp(parser, fmt.Sprintf("%v", errors.Errorf("%v", err)))
	}
	if err := parser.Parse(os.Args[1:]); err != nil {
		//args.Print()
		if args.Version {
			ShowVersion()
			return
		}
		if err.Error() == "error processing default value for input: cannot parse into []string" {
			ShowHelp(parser, "Error: 引数にxlsxファイルを与えてください。")
		}
		if err.Error() == "help requested by user" {
			//ShowHelp(parser, fmt.Sprintf("%v", errors.Errorf("%v", err)))
			ShowHelp(parser, "")
		} else if err.Error() == "version requested by user" {
			ShowVersion()
		} else {
			panic(errors.Errorf("%v", err))
		}
	}
	//if args.Version || args.VersionSub != nil {
	if args.Version {
		ShowVersion()
		return
	}
	//if len(args.Csv) == 0 {
	//	ShowHelp(parser, "Error: 入力csvファイルを指定してください。")
	//}
	//if len(args.Grep) == 0 && args.Row == -1 && args.Col == -1 {
	//	ShowHelp(parser, "Error: 行番号、列番号等を指定してください。")
	//}
	//if args.Debug {
	//	args.Print()
	//}
}
func ShowHelp(parser *arg.Parser, post string) {
	buf := new(bytes.Buffer)
	parser.WriteHelp(buf)
	fmt.Printf("%v\n", strings.ReplaceAll(buf.String(), "display this help and exit", "ヘルプを出力する。"))
	if len(post) != 0 {
		fmt.Println(post)
	}
	os.Exit(1)
}
func ShowVersion() {
	fmt.Printf("%v version %v.%v\n", GetFileNameWithoutExt(os.Args[0]), Version, Revision)
}

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
	return filepath.Base(path[:len(path)-len(filepath.Ext(path))])
}
func GetFilePathWithoutExt(path string) string {
	return filepath.ToSlash(filepath.Join(filepath.Dir(path), GetFileNameWithoutExt(path)))
}
