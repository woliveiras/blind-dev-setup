package target

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func decodeCandidates(output []byte) ([]Candidate, error) {
	var candidates []Candidate
	if err := json.Unmarshal(output, &candidates); err != nil {
		return nil, fmt.Errorf("interpretar resposta do Windows: %w", err)
	}
	for index := range candidates {
		candidates[index].Model = strings.TrimSpace(candidates[index].Model)
		candidates[index].DriveLetter = strings.ToUpper(strings.TrimSpace(
			strings.TrimSuffix(candidates[index].DriveLetter, ":"),
		))
		candidates[index].Label = strings.TrimSpace(candidates[index].Label)
		candidates[index].FileSystem = strings.ToUpper(strings.TrimSpace(candidates[index].FileSystem))
	}
	sort.SliceStable(candidates, func(left, right int) bool {
		if candidates[left].DiskNumber == candidates[right].DiskNumber {
			return candidates[left].DriveLetter < candidates[right].DriveLetter
		}
		return candidates[left].DiskNumber < candidates[right].DiskNumber
	})
	return candidates, nil
}
