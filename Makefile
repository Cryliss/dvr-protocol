build:
	go mod download
	go build --race -o dvr bin/dvr/main.go

run1:
	./dvr -t /Users/sabra/go/src/dvr-protocol/topology/config/topology1.txt -i 60 -d false

run2:
	./dvr -t /Users/sabra/go/src/dvr-protocol/topology/config/topology2.txt -i 60 -d false

run3:
	./dvr -t /Users/sabra/go/src/dvr-protocol/topology/config/topology3.txt -i 60 -d false

run4:
	./dvr -t /Users/sabra/go/src/dvr-protocol/topology/config/topology4.txt -i 60 -d false
