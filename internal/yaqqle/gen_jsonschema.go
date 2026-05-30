package yaqqle

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type jsonSchemaDoc struct {
	Schema     string                   `json:"$schema"`
	Type       string                   `json:"type"`
	Properties map[string]tableProp     `json:"properties"`
}

type tableProp struct {
	Type       string                  `json:"type"`
	Items      *itemProp               `json:"items,omitempty"`
}

type itemProp struct {
	Type       string                    `json:"type"`
	Properties map[string]columnProp     `json:"properties"`
}

type columnProp struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

func GenerateJSONSchema(s *Schema, path string) error {
	doc := jsonSchemaDoc{
		Schema:     "https://json-schema.org/draft-07/schema#",
		Type:       "object",
		Properties: make(map[string]tableProp),
	}

	for _, t := range s.Tables {
		cols := make(map[string]columnProp)
		for _, c := range t.Columns {
			cols[c.EffectiveJSONKey()] = columnProp{
				Type:        pgToJSONSchemaType(c.Type),
				Description: c.Comment,
			}
		}

		doc.Properties[t.Name] = tableProp{
			Type: "array",
			Items: &itemProp{
				Type:       "object",
				Properties: cols,
			},
		}
	}

	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json schema: %w", err)
	}
	return os.WriteFile(path, b, 0644)
}

func pgToJSONSchemaType(pgType string) string {
	m := map[string]string{
		"text":              "string",
		"character varying": "string",
		"char":              "string",
		"integer":           "integer",
		"int":               "integer",
		"bigint":            "integer",
		"smallint":          "integer",
		"boolean":           "boolean",
		"bool":              "boolean",
		"numeric":           "number",
		"real":              "number",
		"double precision":  "number",
		"uuid":              "string",
		"timestamptz":       "string",
		"timestamp":         "string",
		"date":              "string",
		"jsonb":             "object",
		"json":              "object",
	}
	if v, ok := m[strings.ToLower(pgType)]; ok {
		return v
	}
	return "string"
}
