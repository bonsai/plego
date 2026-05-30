package yaqqle

import (
	"fmt"
	"strings"
)

func PostgresDDL(t *Table) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", quoteIdent(t.Name)))

	var pkCols []string
	var lines []string
	for _, c := range t.Columns {
		parts := []string{quoteIdent(c.Name), c.Type}
		if !c.Nullable {
			parts = append(parts, "NOT NULL")
		}
		if c.Default != "" {
			parts = append(parts, "DEFAULT "+c.Default)
		}
		if c.Unique {
			parts = append(parts, "UNIQUE")
		}
		if c.PK {
			pkCols = append(pkCols, quoteIdent(c.Name))
		}
		lines = append(lines, "  "+strings.Join(parts, " "))
	}

	if len(pkCols) > 0 {
		lines = append(lines, fmt.Sprintf("  PRIMARY KEY (%s)", strings.Join(pkCols, ", ")))
	}

	b.WriteString(strings.Join(lines, ",\n"))
	b.WriteString("\n);\n")

	for _, idx := range t.Indexes {
		uniq := ""
		if idx.Unique {
			uniq = "UNIQUE "
		}
		cols := make([]string, len(idx.Columns))
		for i, c := range idx.Columns {
			cols[i] = quoteIdent(c)
		}
		b.WriteString(fmt.Sprintf("CREATE %sINDEX %s ON %s (%s);\n",
			uniq, quoteIdent(idx.Name), quoteIdent(t.Name), strings.Join(cols, ", ")))
	}

	return b.String()
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
