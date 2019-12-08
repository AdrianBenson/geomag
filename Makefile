
GO_PROGS = wsgeomag slgeomag msgeomag

C_LIBS = libmseed libslink

$(C_LIBS):
	$(MAKE) clean all -C vendor/github.com/GeoNet/kit/cvendor/$@

# libslink:
	# $(MAKE) clean all -C vendor/github.com/GeoNet/kit/cvendor/libslink

$(GO_PROGS) : $(C_LIBS)
	cd cmd/$@ && go build -v
	# cd cmd/$@ && go build -v && go install

all : $(GO_PROGS)

