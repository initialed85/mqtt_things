package circumstances_engine

import (
	"fmt"
	"log"
	"strings"
	"time"
)

type Circumstances struct {
	Timestamp                                                                           time.Time
	BeforeSunrise, AfterSunrise, BeforeSunset, AfterSunset, BeforeBedtime, AfterBedtime bool
	Hot, Comfortable, Cold                                                              bool
}

func getTime(timestamp string) (time.Time, error) {
	return time.Parse("15:04:05", timestamp)
}

func recreateForNow(timestamp, now time.Time) time.Time {
	nowDate := now.Format("2006-01-02")
	timestampTime := timestamp.Format("15:04:05")

	recreated, _ := time.Parse(
		"2006-01-02 15:04:05 MST",
		fmt.Sprintf("%v %v %v", nowDate, timestampTime, now.Format("MST")),
	)

	return recreated
}

func CalculateCircumstances(now, sunrise, sunset, bedtime time.Time, temperature, hotEntry, hotExit, coldEntry, coldExit float64, offset time.Duration) Circumstances {
	var beforeSunrise, afterSunrise, beforeSunset, afterSunset, beforeBedtime, afterBedtime, hot, comfortable, cold bool

	log.Printf(
		"getting circumstances for %v, %v, %v, %v, %v, %v, %v, %v, %v, %v",
		now,
		sunrise,
		sunset,
		bedtime,
		temperature,
		hotEntry,
		hotExit,
		coldEntry,
		coldExit,
		offset,
	)

	midnight, _ := getTime("00:00:00")
	midnight = recreateForNow(midnight, now)

	midday, _ := getTime("12:00:00")
	midday = recreateForNow(midday, now)

	afterMidnight := now.After(midnight.Add(-offset)) && now.Before(midday.Add(-offset))
	log.Printf("afterMidnight is %v", afterMidnight)

	if afterMidnight {
		sunrise = recreateForNow(sunrise, now)
		sunset = recreateForNow(sunset, now).Add(-time.Hour * 24)
		bedtime = recreateForNow(bedtime, now).Add(-time.Hour * 24)
	} else {
		sunrise = recreateForNow(sunrise, now).Add(time.Hour * 24)
		sunset = recreateForNow(sunset, now)
		bedtime = recreateForNow(bedtime, now)
	}

	log.Printf(
		"sunrise %v, sunset %v, bedtime %v",
		sunrise,
		sunset,
		bedtime,
	)

	beforeSunrise = now.Before(sunrise.Add(-offset))
	afterSunrise = now.After(sunrise.Add(-offset))

	beforeSunset = now.Before(sunset.Add(-offset))
	afterSunset = now.After(sunset.Add(-offset))

	beforeBedtime = now.Before(bedtime.Add(-offset))
	afterBedtime = now.After(bedtime.Add(-offset))

	cold = temperature <= coldEntry
	comfortable = temperature >= coldExit && temperature <= hotExit
	hot = temperature >= hotEntry

	circumstances := Circumstances{
		Timestamp:     now,
		BeforeSunrise: beforeSunrise,
		AfterSunrise:  afterSunrise,
		BeforeSunset:  beforeSunset,
		AfterSunset:   afterSunset,
		BeforeBedtime: beforeBedtime,
		AfterBedtime:  afterBedtime,
		Hot:           hot,
		Comfortable:   comfortable,
		Cold:          cold,
	}

	log.Printf("circumstances are %+v", circumstances)

	return circumstances
}

type TopicAndCircumstance struct {
	Topic        string
	Circumstance string
}

func convertCircumstance(circumstance bool) string {
	if circumstance {
		return "1"
	} else {
		return "0"
	}
}

func GetTopicsAndCircumstances(circumstances Circumstances, prefix, suffix string) []TopicAndCircumstance {
	prefix = strings.TrimRight(prefix, "/")
	if strings.TrimSpace(suffix) != "" {
		suffix = fmt.Sprintf("_%v", strings.TrimLeft(suffix, "_"))
	}

	return []TopicAndCircumstance{
		{fmt.Sprintf("%v/before_sunrise%v/get", prefix, suffix), convertCircumstance(circumstances.BeforeSunrise)},
		{fmt.Sprintf("%v/after_sunrise%v/get", prefix, suffix), convertCircumstance(circumstances.AfterSunrise)},
		{fmt.Sprintf("%v/before_sunset%v/get", prefix, suffix), convertCircumstance(circumstances.BeforeSunset)},
		{fmt.Sprintf("%v/after_sunset%v/get", prefix, suffix), convertCircumstance(circumstances.AfterSunset)},
		{fmt.Sprintf("%v/before_bedtime%v/get", prefix, suffix), convertCircumstance(circumstances.BeforeBedtime)},
		{fmt.Sprintf("%v/after_bedtime%v/get", prefix, suffix), convertCircumstance(circumstances.AfterBedtime)},
		{fmt.Sprintf("%v/hot%v/get", prefix, suffix), convertCircumstance(circumstances.Hot)},
		{fmt.Sprintf("%v/comfortable%v/get", prefix, suffix), convertCircumstance(circumstances.Comfortable)},
		{fmt.Sprintf("%v/cold%v/get", prefix, suffix), convertCircumstance(circumstances.Cold)},
	}
}
