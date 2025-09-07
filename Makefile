default: out/nikola-telemetry
.PHONY: clean default run

GO_FILES = $(shell find . -name '*.go')
PROTO_FILES = $(shell find proto -name '*.proto')

out:
	mkdir out

out/nikola-telemetry: out gen/.sentinel $(GO_FILES)
	go build -o out .

gen/.sentinel: $(PROTO_FILES)
	buf generate
	@touch gen/.sentinel

clean:
	rm -r out
	rm -r gen
