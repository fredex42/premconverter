BINDIR=../bin
SOURCEDIR=./src

all: premconverter.macos premconverter.linux64 premconverter.linux32 premconverter.windows
clean:
	rm -rf $(BINDIR)/premconverter.*

premconverter.macos: $(SOURCEDIR)/main.go $(SOURCEDIR)/reader/reader.go
	cd $(SOURCEDIR); GOOS=darwin go build -o $(BINDIR)/premconverter.macos

premconverter.linux64: $(SOURCEDIR)/main.go $(SOURCEDIR)/reader/reader.go
	cd $(SOURCEDIR); GOOS=linux GOARCH=amd64 go build -o $(BINDIR)/premconverter.linux64

premconverter.linux32: $(SOURCEDIR)/main.go $(SOURCEDIR)/reader/reader.go
	cd $(SOURCEDIR); GOOS=linux GOARCH=386 go build -o $(BINDIR)/premconverter.linux32

premconverter.windows: $(SOURCEDIR)/main.go $(SOURCEDIR)/reader/reader.go
	cd $(SOURCEDIR); GOOS=windows go build -o $(BINDIR)/premconverter.exe
