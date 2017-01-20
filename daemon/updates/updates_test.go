package updates

import (
	"testing"
	"time"
)

type nextCheckExample struct {
	now       time.Time
	lastCheck time.Time
	expected  time.Duration
}

type mockTimeManager struct {
	now time.Time
}

func (m mockTimeManager) Now() time.Time {
	return m.now
}

func TestNextCheck(t *testing.T) {
	wednesday := time.Time{}
	daysDiff := int(releaseDay) - int(wednesday.Weekday())
	wednesday = wednesday.Add(time.Duration(daysDiff*24-wednesday.Hour()) * time.Hour)

	sunday := wednesday.Add(-24 * 3 * time.Hour)
	tuesday := wednesday.Add(-24 * time.Hour)
	thursday := wednesday.Add(24 * time.Hour)

	examples := []nextCheckExample{
		{
			lastCheck: time.Time{},
			expected:  minCheckDuration,
		},

		{
			now:       tuesday,
			lastCheck: tuesday,
			expected:  30 * time.Hour,
		},

		{
			now:       sunday,
			lastCheck: wednesday.Add(time.Duration(-24*7+releaseHourCheck) * time.Hour),
			expected:  time.Duration(24*3*time.Hour + releaseHourCheck*time.Hour),
		},

		{
			now:       tuesday,
			lastCheck: wednesday.Add(-24 * 7 * time.Hour),
			expected:  minCheckDuration,
		},

		{
			now:       thursday,
			lastCheck: wednesday.Add(releaseHourCheck * time.Hour),
			expected:  time.Duration(24*6+releaseHourCheck) * time.Hour,
		},

		{
			now:       wednesday,
			lastCheck: wednesday.Add(-24 * 7 * time.Hour),
			expected:  minCheckDuration,
		},

		{
			now:       wednesday,
			lastCheck: wednesday,
			expected:  time.Duration(releaseHourCheck * time.Hour),
		},

		{
			now:       wednesday.Add((releaseHourCheck + 1) * time.Hour),
			lastCheck: wednesday.Add(releaseHourCheck * time.Hour),
			expected:  (24*7 - 1) * time.Hour,
		},
	}

	for _, example := range examples {
		timeManager := mockTimeManager{
			now: example.now,
		}

		e := &Engine{
			timeManager: timeManager,
		}

		e.lastCheck = example.lastCheck
		if d := e.nextCheck(); d != example.expected {
			t.Fatalf("expected %s, got %s", example.expected, d)
		}
	}
}
