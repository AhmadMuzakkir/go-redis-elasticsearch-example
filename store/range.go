package store

type Range int

const (
	OneMinute Range = iota
	FiveMinute
	OneHour
	OneDay
	OneWeek
	OneMonth
	NumRange // Used to get all the range constants.
)

// TODO add unit test
// Get the description for the interval
func RangeDescription(rang Range) string {
	switch rang {
	case OneMinute:
		return "1 minute ago"
	case FiveMinute:
		return "5 minutes ago"
	case OneHour:
		return "1 hour ago"
	case OneDay:
		return "1 day ago"
	case OneWeek:
		return "2 week ago"
	case OneMonth:
		return "1 month ago"
	default:
		return ""
	}
}
