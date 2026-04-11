package human_test

import (
	"testing"
	"time"

	"github.com/gechr/primer/human"
	"github.com/stretchr/testify/require"
)

func TestFormatTimeAgoFromNow(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "now", human.FormatTimeAgoFrom(now, now))
}

func TestFormatTimeAgoFromSeconds(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "now", human.FormatTimeAgoFrom(now.Add(-30*time.Second), now))
}

func TestFormatTimeAgoFromOneMinute(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "1 minute ago", human.FormatTimeAgoFrom(now.Add(-1*time.Minute), now))
}

func TestFormatTimeAgoFromMinutes(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "15 minutes ago", human.FormatTimeAgoFrom(now.Add(-15*time.Minute), now))
}

func TestFormatTimeAgoFromOneHour(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "1 hour ago", human.FormatTimeAgoFrom(now.Add(-1*time.Hour), now))
}

func TestFormatTimeAgoFromHours(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "5 hours ago", human.FormatTimeAgoFrom(now.Add(-5*time.Hour), now))
}

func TestFormatTimeAgoFromOneDay(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "1 day ago", human.FormatTimeAgoFrom(now.Add(-24*time.Hour), now))
}

func TestFormatTimeAgoFromDays(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "3 days ago", human.FormatTimeAgoFrom(now.Add(-3*24*time.Hour), now))
}

func TestFormatTimeAgoFromOneWeek(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "1 week ago", human.FormatTimeAgoFrom(now.Add(-7*24*time.Hour), now))
}

func TestFormatTimeAgoFromWeeks(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "3 weeks ago", human.FormatTimeAgoFrom(now.Add(-21*24*time.Hour), now))
}

func TestFormatTimeAgoFromOneMonth(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "1 month ago", human.FormatTimeAgoFrom(now.Add(-35*24*time.Hour), now))
}

func TestFormatTimeAgoFromMonths(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "6 months ago", human.FormatTimeAgoFrom(now.Add(-180*24*time.Hour), now))
}

func TestFormatTimeAgoFromOneYear(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "1 year ago", human.FormatTimeAgoFrom(now.Add(-365*24*time.Hour), now))
}

func TestFormatTimeAgoFromYears(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "2 years ago", human.FormatTimeAgoFrom(now.Add(-730*24*time.Hour), now))
}

func TestFormatTimeAgoFromFuture(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "in 5 hours", human.FormatTimeAgoFrom(now.Add(5*time.Hour), now))
}

func TestFormatTimeAgoCompactFromNow(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "now", human.FormatTimeAgoCompactFrom(now, now))
}

func TestFormatTimeAgoCompactFromMinutes(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "15m ago", human.FormatTimeAgoCompactFrom(now.Add(-15*time.Minute), now))
}

func TestFormatTimeAgoCompactFromHours(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "5h ago", human.FormatTimeAgoCompactFrom(now.Add(-5*time.Hour), now))
}

func TestFormatTimeAgoCompactFromDays(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "3d ago", human.FormatTimeAgoCompactFrom(now.Add(-3*24*time.Hour), now))
}

func TestFormatTimeAgoCompactFromWeeks(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "2w ago", human.FormatTimeAgoCompactFrom(now.Add(-14*24*time.Hour), now))
}

func TestFormatTimeAgoCompactFromMonths(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "6mo ago", human.FormatTimeAgoCompactFrom(now.Add(-180*24*time.Hour), now))
}

func TestFormatTimeAgoCompactFromYears(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "2y ago", human.FormatTimeAgoCompactFrom(now.Add(-730*24*time.Hour), now))
}

func TestFormatTimeAgoCompactFromFuture(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	require.Equal(t, "in 3d", human.FormatTimeAgoCompactFrom(now.Add(3*24*time.Hour), now))
}
