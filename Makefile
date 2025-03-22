
PACKAGES := $$(go list ./...)

test-unit:
	POLARIS_UNIT_TEST_ONLY=true go test -tags ci -v $(PACKAGES) -count=1
