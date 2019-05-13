out/pycmdpev.so: $(wildcard pybindings/*)
	(cd pybindings && go build -buildmode=c-shared -o $(abspath $@) .)

.PHONY: python3
python3: out/pycmdpev.so

.PHONY: install
install:
	go install .
