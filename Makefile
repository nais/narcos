.PHONY: build check staticcheck vulncheck deadcode fmt test vet

build: check fmt test
	go build -o narc cmd/narc/main.go

check: staticcheck vulncheck deadcode vet

staticcheck:
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...

vulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

deadcode:
	go run golang.org/x/tools/cmd/deadcode@latest -test ./...

vet:
	go vet ./...

fmt:
	go run mvdan.cc/gofumpt@latest -w ./

test:
	go test ./...
