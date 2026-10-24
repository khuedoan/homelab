package benchmarks

import (
	"encoding/json"
	"errors"
	"strings"
)

func decodeReport(output string, report any) (string, error) {
	offset := 0
	for _, line := range strings.SplitAfter(output, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "{") {
			return output[:offset], json.Unmarshal([]byte(output[offset:]), report)
		}
		offset += len(line)
	}

	return output, errors.New("benchmark output contains no JSON object")
}
