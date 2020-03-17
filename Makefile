out/gocmdpev: go.mod go.sum main.go $(wildcard pev/*)
	go build -trimpath -o $(abspath $@) .

out/pycmdpev.so: $(wildcard pybindings/*)
	(cd pybindings && go build -buildmode=c-shared -o $(abspath $@) .)

.PHONY: python3
python3: out/pycmdpev.so

.PHONY: install
install:
	go install .
