package reporter

import (
	"encoding/json"
	"os"
)

// JSONReporter produces a JSON data file with all the benchmark information.
type JSONReporter struct {
	outputPath string
}

func (j *JSONReporter) ProduceReport(reports []*BenchmarkReport) error {
	result, err := json.Marshal(reports)
	if err != nil {
		return err
	}

	return os.WriteFile(j.outputPath, []byte(result), 0o600)
}

func NewJSONReporter(output string) *JSONReporter {
	return &JSONReporter{outputPath: output}
}
