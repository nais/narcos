.PHONY: build check staticcheck vulncheck deadcode fmt test

build: check fmt
	go build -o narc cmd/narc/main.go

check: staticcheck vulncheck deadcode

staticcheck:
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...

vulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

deadcode:
	go run golang.org/x/tools/cmd/deadcode@latest -test ./...

fmt:
	go run mvdan.cc/gofumpt@latest -w ./

test:
	go test ./...
