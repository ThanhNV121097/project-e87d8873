package orders

import (
	"fmt"
	"time"
)

// vietnam is the zone the shop's day is counted in. Vietnam has had no
// daylight saving since 1975, so a fixed +07:00 gives the same date as
// Asia/Ho_Chi_Minh for every date an order can have, and needs no zone
// database.
var vietnam = time.FixedZone("ICT", 7*3600)

// Day is the instant seen in Vietnam time: its date is the day of the order
// number, and 18:00 UTC on the 16th is the 17th.
func Day(instant time.Time) time.Time { return instant.In(vietnam) }

// Number is DEO-YYMMDD-NNNN: the date in Vietnam and the day's counter.
func Number(day time.Time, seq int) string {
	return fmt.Sprintf("DEO-%s-%04d", day.Format("060102"), seq)
}

// placedLayout renders an instant with its offset, which prints +07:00.
const placedLayout = "2006-01-02T15:04:05-07:00"
