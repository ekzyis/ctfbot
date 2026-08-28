.PHONY: build test vet clean

build:
	go build -o ctfbot .

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f ctfbot result
