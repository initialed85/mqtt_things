package circumstances_engine

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getTimeForTesting(timeString string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", timeString)
	if err != nil {
		log.Fatal(err)
	}

	return t
}

func TestCalculateCircumstances_Times(t *testing.T) {
	circumstances := Circumstances{}

	// sunrise/sunset hasn't updated yet
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-06 01:00:00"),
		getTimeForTesting("1991-02-05 06:00:00"),
		getTimeForTesting("1991-02-05 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		23, 29, 27, 10, 12, time.Duration(15)*time.Minute,
	)
	assert.Equal(t, false, circumstances.AfterSunrise)
	assert.Equal(t, true, circumstances.AfterSunset)
	assert.Equal(t, true, circumstances.AfterBedtime)

	// sunrise/sunset has updated now
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-06 05:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		23, 29, 27, 10, 12, time.Duration(15)*time.Minute,
	)
	assert.Equal(t, false, circumstances.AfterSunrise)
	assert.Equal(t, true, circumstances.AfterSunset)
	assert.Equal(t, true, circumstances.AfterBedtime)

	// clearly after sunrise
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-06 07:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		23, 29, 27, 10, 12, time.Duration(0),
	)
	assert.Equal(t, true, circumstances.AfterSunrise)
	assert.Equal(t, false, circumstances.AfterSunset)
	assert.Equal(t, false, circumstances.AfterBedtime)

	// clearly after sunrise
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-06 13:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		23, 29, 27, 10, 12, time.Duration(0),
	)
	assert.Equal(t, true, circumstances.AfterSunrise)
	assert.Equal(t, false, circumstances.AfterSunset)
	assert.Equal(t, false, circumstances.AfterBedtime)

	// clearly before sunset
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-06 17:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		23, 29, 27, 10, 12, time.Duration(0),
	)
	assert.Equal(t, true, circumstances.AfterSunrise)
	assert.Equal(t, false, circumstances.AfterSunset)
	assert.Equal(t, false, circumstances.AfterBedtime)

	// clearly after sunset
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-06 19:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		23, 29, 27, 10, 12, time.Duration(0),
	)
	assert.Equal(t, false, circumstances.AfterSunrise)
	assert.Equal(t, true, circumstances.AfterSunset)
	assert.Equal(t, false, circumstances.AfterBedtime)

	// clearly after bedtime
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-06 23:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		23, 29, 27, 10, 12, time.Duration(0),
	)
	assert.Equal(t, false, circumstances.AfterSunrise)
	assert.Equal(t, true, circumstances.AfterSunset)
	assert.Equal(t, true, circumstances.AfterBedtime)
}

func TestCalculateCircumstances_Offset(t *testing.T) {
	circumstances := Circumstances{}

	// before sunrise with offset
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-06 06:14:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		23,
		29,
		27,
		10,
		12,
		time.Duration(15)*time.Minute,
	)
	assert.Equal(t, false, circumstances.AfterSunrise)

	// after sunrise with offset
	circumstances = CalculateCircumstances(
		getTimeForTesting("1991-02-06 06:16:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		getTimeForTesting("1991-02-06 18:00:00"),
		getTimeForTesting("1991-02-06 22:00:00"),
		getTimeForTesting("1991-02-06 06:00:00"),
		23,
		29,
		27,
		10,
		12,
		time.Duration(15)*time.Minute,
	)
	assert.Equal(t, true, circumstances.AfterSunrise)
}

func TestCalculateCircumstances_Temperature(t *testing.T) {
	circumstances := Circumstances{}

	// hot
	circumstances = CalculateCircumstances(
		time.Now(), time.Now(), time.Now(), time.Now(), time.Now(),
		29,
		29, 27, 12, 14, time.Duration(0),
	)
	assert.Equal(t, true, circumstances.Hot)
	assert.Equal(t, false, circumstances.Comfortable)
	assert.Equal(t, false, circumstances.Cold)

	// on the way to hot
	circumstances = CalculateCircumstances(
		time.Now(), time.Now(), time.Now(), time.Now(), time.Now(),
		28,
		29, 27, 12, 14,
		time.Duration(0),
	)
	assert.Equal(t, false, circumstances.Hot)
	assert.Equal(t, false, circumstances.Comfortable)
	assert.Equal(t, false, circumstances.Cold)

	// comfortable
	circumstances = CalculateCircumstances(
		time.Now(), time.Now(), time.Now(), time.Now(), time.Now(),
		23,
		29, 27, 12, 14,
		time.Duration(0),
	)
	assert.Equal(t, false, circumstances.Hot)
	assert.Equal(t, true, circumstances.Comfortable)
	assert.Equal(t, false, circumstances.Cold)

	// on the way to cold
	circumstances = CalculateCircumstances(
		time.Now(), time.Now(), time.Now(), time.Now(), time.Now(),
		13,
		29, 27, 12, 14,
		time.Duration(0),
	)
	assert.Equal(t, false, circumstances.Hot)
	assert.Equal(t, false, circumstances.Comfortable)
	assert.Equal(t, false, circumstances.Cold)

	// cold
	circumstances = CalculateCircumstances(
		time.Now(), time.Now(), time.Now(), time.Now(), time.Now(),
		12,
		29, 27, 12, 14,
		time.Duration(0),
	)
	assert.Equal(t, false, circumstances.Hot)
	assert.Equal(t, false, circumstances.Comfortable)
	assert.Equal(t, true, circumstances.Cold)
}

func TestGetTopicsAndCircumstances(t *testing.T) {
	circumstances := Circumstances{
		time.Now(),
		false,
		true,
		true,
		false,
		true,
		false,
		false,
	}

	assert.Equal(t,
		[]TopicAndCircumstance{
			{Topic: "home/circumstances/after_sunrise_15m_later/get", Circumstance: "0"},
			{Topic: "home/circumstances/after_sunset_15m_later/get", Circumstance: "1"},
			{Topic: "home/circumstances/after_bedtime_15m_later/get", Circumstance: "1"},
			{Topic: "home/circumstances/after_waketime_15m_later/get", Circumstance: "0"},
			{Topic: "home/circumstances/hot_15m_later/get", Circumstance: "1"},
			{Topic: "home/circumstances/comfortable_15m_later/get", Circumstance: "0"},
			{Topic: "home/circumstances/cold_15m_later/get", Circumstance: "0"}},
		GetTopicsAndCircumstances(circumstances, "home/circumstances", "_15m_later"),
	)
}
