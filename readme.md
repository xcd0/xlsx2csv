# xlsx2csv

引数にxlsxを与えると、csvに変換して出力する。

複数シートがあるとき、複数のcsvとして出力する。

```sh
$ ls
xlsx2csv.exe  sample.xlsx
$ ./xlsx2csv.exe sample.xlsx
$ ls
xlsx2csv.exe  sample.xlsx  sample_0_Sheet1.csv
```

