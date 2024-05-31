BIN           := ./xlsx2csv.exe
REVISION      := `git rev-parse --short HEAD`
FLAG          :=  -a -tags netgo -trimpath -ldflags='-s -w -extldflags="-static" -buildid='
all:
	cat ./makefile
build:
	make clean
	go build
release:
	make clean
	go build $(FLAG)
	make upx 
	@echo Success!
upx:
	upx --lzma $(BIN)
clean:
	rm -rf *.csv embedded_files.go
	go generate
	goimports -w *.go
	gofmt -w *.go

