package yaqqle

import (
	"encoding/json"
	"fmt"
	"os"
)

func GenerateJSON(s *Schema, path string) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	return os.WriteFile(path, b, 0644)
}
