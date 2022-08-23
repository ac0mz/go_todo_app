package clock

import "time"

type Clocker interface {
	Now() time.Time
}

// RealClocker は現在の時刻情報を扱う
type RealClocker struct{}

func (rc RealClocker) Now() time.Time {
	return time.Now()
}

// FixedClocker は固定化された時刻情報を扱う
type FixedClocker struct{}

func (fc FixedClocker) Now() time.Time {
	return time.Date(2022, 8, 23, 23, 59, 59, 0, time.UTC)
}
