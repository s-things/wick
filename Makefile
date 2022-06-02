deps:
	go get github.com/s-things/wick/cmd/wick

build:
	go build github.com/s-things/wick/cmd/wick

test:
	go test github.com/s-things/wick/cmd/wick -v
	go test github.com/s-things/wick/core -v

run:
	./wick

clean:
	rm -f wick
