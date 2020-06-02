out/gocmdpev: go.mod go.sum main.go $(wildcard pev/*)
	go build -trimpath -o $(abspath $@) .

out/pycmdpev.so: $(wildcard pybindings/*)
	(cd pybindings && go build -buildmode=c-shared -o $(abspath $@) .)

.PHONY: python3
python3: out/pycmdpev.so
	(cd out; python3 -c 'import pycmdpev, sys; pycmdpev.visualize(open(sys.argv[1]).read())' $(abspath example.json))

.PHONY: test
test: out/gocmdpev
	cat example.json | out/gocmdpev

.PHONY: install
install:
	go install .
