build:
	go build -o ./bin/report-generator

build-verbose:
	go build -o ./bin/report-generator -x

clean:
	rm -r ./bin
	
rebuild:
	make clean && make build

run:
	make rebuild && ./bin/report-generator
