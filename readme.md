# xlsx2csv

引数にxlsxを取り、csvに変換して出力する。
複数シートがある場合、複数ファイルのcsvとして出力する。

## usage

```sh
$ ls
xlsx2csv.exe  sample.xlsx
$ ./xlsx2csv.exe sample.xlsx
$ ls
xlsx2csv.exe  sample.xlsx  sample_0_Sheet1.csv
```

## install

```sh
go install github.com/xcd0/xlsx2csv@latest
```

