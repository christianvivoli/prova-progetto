package query

import "fmt"

// FormatLimitOffset returns a SQL string for a given limit & offset.
// Clauses are only added if limit and/or offset are greater than zero.
func FormatLimitOffset(limit, offset int) string {
	if limit > 0 && offset > 0 {
		return fmt.Sprintf(`LIMIT %d OFFSET %d`, limit, offset)
	} else if limit > 0 {
		return fmt.Sprintf(`LIMIT %d`, limit)
	} else if offset > 0 {
		return fmt.Sprintf(`OFFSET %d`, offset)
	}
	return ""
}

// FormatLimitPage returns a SQL string for a given limit & page.
func FormatLimitPage(limit, page int) string {
	if page == 0 {
		page = 1
	}
	offset := (page - 1) * limit
	return FormatLimitOffset(limit, offset)
}
