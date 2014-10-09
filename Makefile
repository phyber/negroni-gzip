PKGNAME=github.com/phyber/negroni-gzip/gzip
GOCMD=go
GOTOOL=$(GOCMD) tool
GOTEST=$(GOCMD) test
COVERFILE=cover.out
TESTCOVER=$(GOTEST) -coverprofile $(COVERFILE)
GOCOVER=$(GOTOOL) cover -func=$(COVERFILE)

test:
	$(GOTEST) $(PKGNAME)

testcover:
	$(TESTCOVER) $(PKGNAME)
	$(GOCOVER)
