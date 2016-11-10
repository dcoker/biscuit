FROM golang:1.7
ADD GLOCKFILE src/github.com/dcoker/biscuit/
ADD Makefile src/github.com/dcoker/biscuit/
WORKDIR src/github.com/dcoker/biscuit
RUN make glock-sync
ADD . .
RUN make build
