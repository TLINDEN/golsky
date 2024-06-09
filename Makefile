.PHONY: all
all:
	make -C src
	mv src/golsky .

.PHONY: clean
clean:
	make -C src clean
	rm -f dump* rect*
