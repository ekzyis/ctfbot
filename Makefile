.PHONY: build leet test vet clean

build:
	go build -o ctfbot .

leet:
	go build -o leet ./cmd/leet

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f ctfbot leet result
