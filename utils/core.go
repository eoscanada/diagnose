package utils

import bt "cloud.google.com/go/bigtable"

func HasAllColumns(row bt.Row, columns ...string) bool {
	for _, column := range columns {
		if !HasBtColumn(row, column) {
			return false
		}
	}

	return true
}

func HasBtColumn(row bt.Row, familyColumn string) bool {
	for _, cols := range row {
		for _, el := range cols {
			if el.Column == familyColumn {
				return true
			}
		}
	}

	return false
}
