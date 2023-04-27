.PHONY: gen
gen:
	go generate ./...

tests:
	go test -race -v ./... -count=1
