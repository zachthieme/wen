package wen

import "testing"

func TestDateExprInterface(t *testing.T) {
	t.Parallel()
	nodes := []Expr{
		&RelativeDayExpr{},
		&ModWeekdayExpr{},
		&RelativeOffsetExpr{},
		&CountedWeekdayExpr{},
		&OrdinalWeekdayExpr{},
		&LastWeekdayInMonthExpr{},
		&AbsoluteDateExpr{},
		&PeriodRefExpr{},
		&BoundaryExpr{},
		&MultiDateExpr{},
		&WithTimeExpr{},
	}
	for _, n := range nodes {
		n.expr()
	}
}
