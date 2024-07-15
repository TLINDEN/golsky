.PHONY all:
all: build

.PHONY: build
build:
	make -C src
	mv src/golsky .

.PHONY: clean
clean:
	make -C src clean
	rm -f dump* rect*

.PHONY: profile
profile: build
	./golsky -W 1500 -H 1500 -d --profile-file cpu.profile
	go tool pprof --http localhost:8888 golsky cpu.profile
