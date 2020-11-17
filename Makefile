
GO_PROGS = wsgeomag slgeomag msgeomag

C_LIBS = mseed slink

all: $(GO_PROGS)

clean:
	for c in $(C_LIBS); do $(MAKE) -C internal/$$c clean; done ;
	for g in $(GO_PROGS); do (cd cmd/$$g && go clean -v ); done ;

$(C_LIBS):
	$(MAKE) clean all -C internal/$@

$(GO_PROGS) : $(C_LIBS)
	cd cmd/$@ && go build -v

