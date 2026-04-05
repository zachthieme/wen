package wen

import "testing"

func TestDateExprInterface(t *testing.T) {
	t.Parallel()
	nodes := []dateExpr{
		&relativeDayExpr{},
		&modWeekdayExpr{},
		&relativeOffsetExpr{},
		&countedWeekdayExpr{},
		&ordinalWeekdayExpr{},
		&lastWeekdayInMonthExpr{},
		&absoluteDateExpr{},
		&periodRefExpr{},
		&boundaryExpr{},
		&multiDateExpr{},
		&withTimeExpr{},
	}
	for _, n := range nodes {
		n.dateExpr()
	}
}
