build:
	go build -o ./bin/report-generator ./cmd/main.go

clean:
	rm -f test.sqlite
	rm -rf ./bin
	
rebuild:
	make clean && make build

run:
	make rebuild
	./bin/report-generator
