GO=go

PROG=rpz_summarize

main: $(PROG)

$(PROG): $(PROG).go
	GOPATH=`pwd` $(GO) build $(PROG).go

clean:
	rm -f $(PROG)
	rm -f *.o
	rm -f *~
	rm -f \#*
