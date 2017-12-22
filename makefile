USER=$(shell whoami)
HEAD=$(shell git describe --tags 2> /dev/null  || git rev-parse --short HEAD)
STAMP=$(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
DIRTY=$(shell test $(shell git status --porcelain | wc -l) -eq 0 || echo '(dirty)')

LDFLAGS="-X main.buildStamp=$(STAMP) -X main.buildUser=$(USER) -X main.buildHash=$(HEAD) -X main.buildDirty=$(DIRTY)"

build: darwin64 linux64

clean:
	-rm -f sqlread
	-rm -rf release

darwin64:
	env GOOS=darwin GOARCH=amd64 go clean -i
	env GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS) -o release/darwin64/sqlread ./cmd/sqlread

linux64:
	env GOOS=linux GOARCH=amd64 go clean -i
	env GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -o release/linux64/sqlread ./cmd/sqlread

.PHONY: release
release:
	$(MAKE) clean
	$(MAKE) build

	zip release/sqlread.darwin_amd64.$(HEAD).zip release/darwin64/sqlread
	zip release/sqlread.linux_amd64.$(HEAD).zip release/linux64/sqlread