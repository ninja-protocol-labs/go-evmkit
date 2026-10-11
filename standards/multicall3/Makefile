FUZZTIME ?= 30s
PARALLEL ?= 2

.PHONY: fuzz
fuzz:
	@for name in $$(go test -list '^Fuzz' . | grep '^Fuzz'); do \
		( \
			echo "=== $$name start ==="; \
			go test -run '^$$' -fuzz "^$${name}\$$" -fuzztime=$(FUZZTIME) -parallel=$(PARALLEL) . ; \
			echo "=== $$name done ===" \
		) & \
	done; \
	wait
