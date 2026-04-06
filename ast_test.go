package wen

import "testing"

func TestDateExprInterface(t *testing.T) {
	t.Parallel()
	// Verify all expression types satisfy the Expr interface.
	// The assignment to []Expr is the check; compilation fails
	// if any type is missing resolveWith.
	_ = []Expr{
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
}
