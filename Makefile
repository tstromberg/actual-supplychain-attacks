.PHONY: all compile run clean

all: compile run

compile: out/yaml2csv

out/yaml2csv: $(wildcard ./cmd/yaml2csv/*.go)
	@mkdir -p out
	(cd cmd/yaml2csv && go build -o ../../out/yaml2csv .)

run: out/yaml2csv
	(cd oss && ../out/yaml2csv > ../oss_summary.csv)
	(cd proprietary && ../out/yaml2csv > ../proprietary_summary.csv)

clean:
	@echo "Cleaning up generated files..."
	rm -f out/yaml2csv oss_summary.csv proprietary_summary.csv
	@rmdir --ignore-fail-on-non-empty out
