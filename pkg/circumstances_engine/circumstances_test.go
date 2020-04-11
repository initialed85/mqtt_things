package circumstances_engine

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func getTimeForTesting(timeString string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", timeString)
	if err != nil {
		log.Fatal(err)
	}

	return t
}

func TestCalculateCircumstances_Sunrise(t *testing.T) {
	circumstances := Circumstances{}

	// before
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-06 05:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		23,
		29,
		27,
		10,
		12,
		time.Duration(0),
	)
	assert.Equal(t, true, circumstances.BeforeSunrise)
	assert.Equal(t, false, circumstances.AfterSunrise)

	// before (but not according to the data because it hasn't updated for some reason)
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-07 05:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		23,
		29,
		27,
		10,
		12,
		time.Duration(0),
	)
	assert.Equal(t, true, circumstances.BeforeSunrise)
	assert.Equal(t, false, circumstances.AfterSunrise)

	// after
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-06 07:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		23,
		29,
		27,
		10,
		12,
		time.Duration(0),
	)
	assert.Equal(t, false, circumstances.BeforeSunrise)
	assert.Equal(t, true, circumstances.AfterSunrise)

	// after (but not according to the data because it hasn't updated for some reason)
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-07 07:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		23,
		29,
		27,
		10,
		12,
		time.Duration(0),
	)
	assert.Equal(t, false, circumstances.BeforeSunrise)
	assert.Equal(t, true, circumstances.AfterSunrise)

	// after both sunrise and sunset (special case; look to next sunrise, not previous)
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-07 19:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		23,
		29,
		27,
		10,
		12,
		time.Duration(0),
	)
	assert.Equal(t, true, circumstances.BeforeSunrise)
	assert.Equal(t, false, circumstances.AfterSunrise)
}

func TestCalculateCircumstances_Sunset(t *testing.T) {
	circumstances := Circumstances{}

	// before
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-06 17:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		0, 0, 0, 0, 0,
		time.Duration(0),
	)
	assert.Equal(t, true, circumstances.BeforeSunset)
	assert.Equal(t, false, circumstances.AfterSunset)

	// before (but not according to the data because it hasn't updated for some reason)
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-07 17:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		0, 0, 0, 0, 0,
		time.Duration(0),
	)
	assert.Equal(t, true, circumstances.BeforeSunset)
	assert.Equal(t, false, circumstances.AfterSunset)

	// after
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-06 19:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		0, 0, 0, 0, 0,
		time.Duration(0),
	)
	assert.Equal(t, false, circumstances.BeforeSunset)
	assert.Equal(t, true, circumstances.AfterSunset)

	// after (but after midnight)
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-07 01:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-07 22:00:00"),
		0, 0, 0, 0, 0,
		time.Duration(0),
	)
	assert.Equal(t, false, circumstances.BeforeSunset)
	assert.Equal(t, true, circumstances.AfterSunset)

	// after both sunset and bedtime
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-07 23:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		23,
		29,
		27,
		10,
		12,
		time.Duration(0),
	)
	assert.Equal(t, false, circumstances.BeforeSunset)
	assert.Equal(t, true, circumstances.AfterSunset)
}

func TestCalculateCircumstances_Bedtime(t *testing.T) {
	circumstances := Circumstances{}

	// before
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-06 21:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		0, 0, 0, 0, 0,
		time.Duration(0),
	)
	assert.Equal(t, true, circumstances.BeforeBedtime)
	assert.Equal(t, false, circumstances.AfterBedtime)

	// before (but not according to the data because it hasn't updated for some reason)
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-07 21:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		0, 0, 0, 0, 0,
		time.Duration(0),
	)
	assert.Equal(t, true, circumstances.BeforeBedtime)
	assert.Equal(t, false, circumstances.AfterBedtime)

	// after
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-06 23:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		0, 0, 0, 0, 0,
		time.Duration(0),
	)
	assert.Equal(t, false, circumstances.BeforeBedtime)
	assert.Equal(t, true, circumstances.AfterBedtime)

	// after (but not according to the data because it hasn't updated for some reason)
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-07 23:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		0, 0, 0, 0, 0,
		time.Duration(0),
	)
	assert.Equal(t, false, circumstances.BeforeBedtime)
	assert.Equal(t, true, circumstances.AfterBedtime)

	// after (but it's the next day and so now has also been adjusted)
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-07 23:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-07 22:00:00"),
		0, 0, 0, 0, 0,
		time.Duration(0),
	)
	assert.Equal(t, false, circumstances.BeforeBedtime)
	assert.Equal(t, true, circumstances.AfterBedtime)

	// after (but our now is newer than the data)
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-07 23:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		0, 0, 0, 0, 0,
		time.Duration(0),
	)
	assert.Equal(t, false, circumstances.BeforeBedtime)
	assert.Equal(t, true, circumstances.AfterBedtime)
}

func TestCalculateCircumstances_Offset(t *testing.T) {
	circumstances := Circumstances{}

	// before offset sunrise
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-06 05:44:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		23,
		29,
		27,
		10,
		12,
		time.Duration(15)*time.Minute,
	)
	assert.Equal(t, true, circumstances.BeforeSunrise)
	assert.Equal(t, false, circumstances.AfterSunrise)

	// after offset sunrise
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-06 05:46:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		23,
		29,
		27,
		10,
		12,
		time.Duration(15)*time.Minute,
	)
	assert.Equal(t, false, circumstances.BeforeSunrise)
	assert.Equal(t, true, circumstances.AfterSunrise)
}

func TestCalculateCircumstances_Hot(t *testing.T) {
	circumstances := Circumstances{}

	// hot
	circumstances = CalculateCircumstances(
		time.Now(), time.Now(), time.Now(), time.Now(),
		30,
		29,
		27,
		12,
		14,
		time.Duration(0),
	)
	assert.Equal(t, true, circumstances.Hot)
	assert.Equal(t, false, circumstances.Comfortable)

	// hot, but getting colder (or not quite there yet)
	circumstances = CalculateCircumstances(
		time.Now(), time.Now(), time.Now(), time.Now(),
		28,
		29,
		27,
		12,
		14,
		time.Duration(0),
	)
	assert.Equal(t, false, circumstances.Hot)
	assert.Equal(t, false, circumstances.Comfortable)

	// not hot
	circumstances = CalculateCircumstances(
		time.Now(), time.Now(), time.Now(), time.Now(),
		26,
		29,
		27,
		12,
		14,
		time.Duration(0),
	)
	assert.Equal(t, false, circumstances.Hot)
	assert.Equal(t, true, circumstances.Comfortable)
}

func TestCalculateCircumstances_Comfortable(t *testing.T) {
	circumstances := Circumstances{}

	// cold
	circumstances = CalculateCircumstances(
		time.Now(), time.Now(), time.Now(), time.Now(),
		13,
		29,
		27,
		12,
		14,
		time.Duration(0),
	)
	assert.Equal(t, false, circumstances.Comfortable)

	// comfortable
	circumstances = CalculateCircumstances(
		time.Now(), time.Now(), time.Now(), time.Now(),
		23,
		29,
		27,
		12,
		14,
		time.Duration(0),
	)
	assert.Equal(t, true, circumstances.Comfortable)

	// hot
	circumstances = CalculateCircumstances(
		time.Now(), time.Now(), time.Now(), time.Now(),
		28,
		29,
		27,
		12,
		14,
		time.Duration(0),
	)
	assert.Equal(t, false, circumstances.Comfortable)
}

func TestCalculateCircumstances_Cold(t *testing.T) {
	circumstances := Circumstances{}

	// cold
	circumstances = CalculateCircumstances(
		time.Now(), time.Now(), time.Now(), time.Now(),
		11,
		29,
		27,
		12,
		14,
		time.Duration(0),
	)
	assert.Equal(t, true, circumstances.Cold)
	assert.Equal(t, false, circumstances.Comfortable)

	// cold, but getting warmer (or not quite there yet)
	circumstances = CalculateCircumstances(
		time.Now(), time.Now(), time.Now(), time.Now(),
		13,
		29,
		27,
		12,
		14,
		time.Duration(0),
	)
	assert.Equal(t, false, circumstances.Cold)
	assert.Equal(t, false, circumstances.Comfortable)

	// not cold
	circumstances = CalculateCircumstances(
		time.Now(), time.Now(), time.Now(), time.Now(),
		15,
		29,
		27,
		12,
		14,
		time.Duration(0),
	)
	assert.Equal(t, false, circumstances.Cold)
	assert.Equal(t, true, circumstances.Comfortable)
}

func TestGetTopicsAndCircumstances(t *testing.T) {
	circumstances := Circumstances{
		time.Now(),
		true,
		false,
		false,
		false,
		false,
		true,
		false,
		false,
		true,
	}

	assert.Equal(t,
		[]TopicAndCircumstance{
			{"home/circumstances/before_sunrise_15m_later/get", "1"},
			{"home/circumstances/after_sunrise_15m_later/get", "0"},
			{"home/circumstances/before_sunset_15m_later/get", "0"},
			{"home/circumstances/after_sunset_15m_later/get", "0"},
			{"home/circumstances/before_bedtime_15m_later/get", "0"},
			{"home/circumstances/after_bedtime_15m_later/get", "1"},
			{"home/circumstances/hot_15m_later/get", "0"},
			{"home/circumstances/comfortable_15m_later/get", "0"},
			{"home/circumstances/cold_15m_later/get", "1"},
		},
		GetTopicsAndCircumstances(circumstances, "home/circumstances", "_15m_later"),
	)
}
