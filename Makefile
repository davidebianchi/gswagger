VERSION ?= latest

# Create a variable that contains the current date in UTC
# Different flow if this script is running on Darwin or Linux machines.
ifeq (Darwin,$(shell uname))
	NOW_DATE = $(shell date -u +%d-%m-%Y)
else
	NOW_DATE = $(shell date -u -I)
endif

all: test

.PHONY: test
test:
	go test ./... -coverprofile coverage.out
	$(MAKE) clean

.PHONY: version
version:
	sed -i.bck "s|## Unreleased|## ${VERSION} - ${NOW_DATE}|g" "CHANGELOG.md"
	rm -fr "CHANGELOG.md.bck"
	git add "CHANGELOG.md"
	git commit -m "Upgrade version to v${VERSION}"
	git tag v${VERSION}
