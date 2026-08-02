PROFILING ?= 0

.PHONY: gen
gen:
	go generate ./...

tests:
# go test -race -v -gcflags="-N -1" -coverprofile=./cover.out -count=1 ./...
	go test -race -v -coverprofile=./cover.out -count=1 ./...

.PHONY: bench
bench:
# Run benchmark without compiler optimizatio
# go test -v -gcflags="-N -l" -test.run=NONE -bench=. -benchtime=10s -benchmem -memprofile memprofile.out -cpuprofile profile.out -count=1
ifeq ($(PROFILING), 1)
	go test -v -test.run=NONE -bench=. \
		-benchmem -memprofile memprofile.out \
		-cpuprofile profile.out -count=1
else
	go test -v -test.run=NONE -bench=. -count=1
endif
