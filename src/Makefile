BINDIR=../bin

all: premconverter.macos premconverter.linux64 premconverter.linux32 premconverter.windows

premconverter.macos: main.go reader/reader.go
	GOOS=darwin go build -o $(BINDIR)/premconverter.macos

premconverter.linux64: main.go reader/reader.go
	GOOS=linux GOARCH=amd64 go build -o $(BINDIR)/premconverter.linux64

premconverter.linux32: main.go reader/reader.go
	GOOS=linux GOARCH=386 go build -o $(BINDIR)/premconverter.linux32

premconverter.windows: main.go reader/reader.go
	GOOS=windows go build -o $(BINDIR)/premconverter.exe
