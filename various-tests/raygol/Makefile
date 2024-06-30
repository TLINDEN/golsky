CFLAGS = -Wall -Wextra -Werror -O2 -g 
LDFLAGS= -L/usr/local/lib -lraylib -lGL -lm -lpthread -ldl -lrt -lX11 -g
CC     = clang
OBJS   = main.o game.o grid.o
DST    = golsky
PREFIX = /usr/local
UID    = root
GID    = 0
MAN    = udpxd.1

.PHONY: all
all: $(DST)

$(DST): $(OBJS)
	$(CC) $(OBJS) $(LDFLAGS) -o $(DST)

%.o: %.c
	$(CC) -c $(CFLAGS) $*.c -o $*.o

.PHONY: clean
clean:
	rm -f *.o $(DST)

.PHONY: install
install: $(DST)
	install -d -o $(UID) -g $(GID) $(PREFIX)/sbin
	install -d -o $(UID) -g $(GID) $(PREFIX)/man/man1
	install -o $(UID) -g $(GID) -m 555 $(DST) $(PREFIX)/sbin/
	install -o $(UID) -g $(GID) -m 444 $(MAN) $(PREFIX)/man/man1/

.PHONY: run
run:
	LD_LIBRARY_PATH=/usr/local/lib ./golsky
