.PHONY: build check staticcheck vulncheck deadcode fmt test vet

build: check fmt test
	go build -o narc cmd/narc/main.go

check: staticcheck vulncheck deadcode vet

staticcheck:
	go tool honnef.co/go/tools/cmd/staticcheck ./...

vulncheck:
	go tool golang.org/x/vuln/cmd/govulncheck ./...

deadcode:
	go tool golang.org/x/tools/cmd/deadcode -test ./...

vet:
	go vet ./...

fmt:
	go tool mvdan.cc/gofumpt -w ./

test:
	go test ./...
