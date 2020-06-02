out/gocmdpev: go.mod go.sum main.go $(wildcard pev/*)
	go build -trimpath -o $(abspath $@) .

out/pycmdpev.so: $(wildcard pybindings/*)
	(cd pybindings && go build -buildmode=c-shared -o $(abspath $@) .)

.PHONY: python3
python3: out/pycmdpev.so
	pkg-config --exists python3 && (cd out; python3 -c 'import pycmdpev, sys; pycmdpev.visualize(open(sys.argv[1]).read())' $(abspath example.json))

.PHONY: python3-docker
python3-docker:
	echo 'FROM golang:1.14-buster\nRUN apt-get update && apt-get install --yes --no-install-recommends python3-dev' | docker build -t gocmdpev-python3 -
	docker run -v "$(abspath .):/workspace" -w /workspace -it --rm gocmdpev-python3 make out/pycmdpev.so

.PHONY: test
test: out/gocmdpev
	cat example.json | out/gocmdpev

.PHONY: install
install:
	go install .
