.PHONY: build check staticcheck vulncheck deadcode fmt test vet

build: check fmt test
	go build -o narc cmd/narc/main.go

check: staticcheck vulncheck deadcode vet

staticcheck:
	go tool staticcheck ./...

vulncheck:
	go tool govulncheck ./...

deadcode:
	go tool deadcode -test ./...

vet:
	go vet ./...

fmt:
	go tool gofumpt -w ./

test:
	go test ./...
