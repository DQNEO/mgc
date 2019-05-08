# All the commands are supposed to run on Linux.
# I use Docker like below.
# docker run -it --cap-add=SYS_PTRACE --security-opt='seccomp=unconfined' --rm -w /mnt -v `pwd`:/mnt dqneo/ubuntu-build-essential:go bash

all: minigo /tmp/out

/tmp/out:
	mkdir /tmp/out

minigo: *.go internalcode.go stdlib.go
	go build -o minigo *.go

internalcode.go: internalcode/runtime.go
	./cp-internalcode.sh

stdlib.go: stdlib/*/*.go
	./concate-stdlib.sh > stdlib.go

/tmp/out/minigo.s: *.go minigo
	./minigo *.go > /tmp/out/minigo.s

minigo2: /tmp/out/minigo.s # 2nd generation
	gcc -g -no-pie -o minigo2 /tmp/out/minigo.s

test2gen: minigo2
	./minigo2 --version
	./unit_test.sh  minigo2 min 2
	./unit_test.sh  minigo2 hello 2


test1gen: all
	./test1gen.sh

test: all
	make test1gen
	make test2gen


circlecitest: all
	make test1gen
	make test2gen

clean:
	rm -f minigo*
	rm -f a.s a.out
	rm -f /tmp/out/*
	rm -f stdlib.go
	rm -f internalcode.go

fmt:
	gofmt -w *.go t/*/*.go
