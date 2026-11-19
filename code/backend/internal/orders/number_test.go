package orders

import (
	"testing"
	"time"
)

func TestTheNumberCountsTheDayInVietnam(t *testing.T) {
	cases := []struct {
		instant string
		seq     int
		want    string
	}{
		{"2026-11-16T02:00:00Z", 1, "DEO-261116-0001"},
		{"2026-11-16T02:00:00Z", 2, "DEO-261116-0002"},
		{"2026-11-16T16:59:59Z", 1, "DEO-261116-0001"},
		{"2026-11-16T17:00:00Z", 1, "DEO-261117-0001"},
		{"2026-11-16T18:00:00Z", 1, "DEO-261117-0001"},
		{"2026-12-31T20:00:00Z", 9999, "DEO-270101-9999"},
	}
	for _, c := range cases {
		instant, err := time.Parse(time.RFC3339, c.instant)
		if err != nil {
			t.Fatal(err)
		}
		if got := Number(Day(instant), c.seq); got != c.want {
			t.Errorf("%s #%d: %s, want %s", c.instant, c.seq, got, c.want)
		}
	}
}

func TestPlacedIsWrittenWithTheVietnamOffset(t *testing.T) {
	instant := time.Date(2026, 11, 16, 18, 0, 0, 0, time.UTC)
	if got := Day(instant).Format(placedLayout); got != "2026-11-17T01:00:00+07:00" {
		t.Errorf("placed = %s, want 2026-11-17T01:00:00+07:00", got)
	}
}
