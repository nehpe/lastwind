BINARIES = lastwind forecast
COVER_PROFILE = coverage.out

.PHONY: all build test cover cover-html clean

all: build

build:
	go build -o lastwind ./cmd/lastwind/
	go build -o forecast ./cmd/forecast/

test:
	go test ./... -v -count=1

cover:
	go test ./... -coverprofile=$(COVER_PROFILE) -count=1
	go tool cover -func=$(COVER_PROFILE)

cover-html: cover
	go tool cover -html=$(COVER_PROFILE) -o coverage.html
	@echo "Coverage report: coverage.html"

clean:
	rm -f $(BINARIES) $(COVER_PROFILE) coverage.html
