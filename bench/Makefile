suite := all

.PHONY: bench
bench:
	$(info Running benchmark suite: $(suite))

ifneq (,$(findstring all,$(suite))) # Check if default ALL value is set
	$(eval override suite = ...)
else ifneq ($(wildcard $(suite)/.),) # Check to make sure the requested suite exists
	$(eval override suite = $(suite)/...)
endif

	go test -bench=. ./$(suite)