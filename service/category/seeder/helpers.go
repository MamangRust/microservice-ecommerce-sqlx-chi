package seeder

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func toBoolPtr(b bool) *bool {
	return &b
}

func toInt32Ptr(i int32) *int32 {
	return &i
}

func toFloat64Ptr(f float64) *float64 {
	return &f
}

func toStringPtr(s string) *string {
	return &s
}

func toDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

func toTime(t time.Time) pgtype.Time {
	return pgtype.Time{Microseconds: int64(t.Hour()*3600+t.Minute()*60+t.Second()) * 1e6, Valid: true}
}
