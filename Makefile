default: out/tesla-operator
.PHONY: clean default run

out:
	@mkdir out

out/tesla-operator: out **/*.go
	@go build -o out ./bin/tesla-operator

clean:
	@rm -r out
