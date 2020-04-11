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

func getDate(timestamp time.Time) time.Time {
	t, _ := time.Parse("2006-01-02", timestamp.Format("2006-01-02"))

	return t
}

func CalculateCircumstances(now, sunrise, sunset, bedtime time.Time, temperature, hotEntry, hotExit, coldEntry, coldExit float64, offset time.Duration) Circumstances {
	var beforeSunrise, afterSunrise, beforeSunset, afterSunset, beforeBedtime, afterBedtime, hot, comfortable, cold bool

	// move sunrise to same day to handle any missing data
	sunriseDiff := getDate(now).Sub(getDate(sunrise)).Hours() / 24
	if sunriseDiff > 0 {
		log.Printf("sunrise %v is old by %v days, fixing", sunrise, sunriseDiff)
		sunrise = sunrise.Add(time.Duration(sunriseDiff) * time.Hour * 24)
	}

	// move sunset to same day to handle any missing data
	sunsetDiff := getDate(now).Sub(getDate(sunset)).Hours() / 24
	if sunsetDiff > 0 {
		log.Printf("sunset %v is old by %v days, fixing", sunset, sunsetDiff)
		sunset = sunset.Add(time.Duration(sunsetDiff) * time.Hour * 24)
	}

	// move bedtime to same day to handle any missing data
	bedtimeDiff := getDate(now).Sub(getDate(bedtime)).Hours() / 24
	if bedtimeDiff > 0 {
		log.Printf("bedtime %v is old by %v days, fixing", bedtime, bedtimeDiff)
		bedtime = bedtime.Add(time.Duration(bedtimeDiff) * time.Hour * 24)
	}

	beforeSunrise = now.Before(sunrise.Add(-offset))
	afterSunrise = now.After(sunrise.Add(-offset))

	beforeSunset = now.Before(sunset.Add(-offset))
	afterSunset = now.After(sunset.Add(-offset))

	beforeBedtime = now.Before(bedtime.Add(-offset))
	afterBedtime = now.After(bedtime.Add(-offset))

	cold = temperature <= coldEntry
	comfortable = temperature >= coldExit && temperature <= hotExit
	hot = temperature >= hotEntry

	// once we're after sunset, look at the next sunrise (not the previous)
	if afterSunset && afterSunrise {
		beforeSunrise = true
		afterSunrise = false
	}

	// once we're after bedtime, look at the next sunset (not the previous)
	if afterBedtime && afterSunset {
		beforeSunset = true
		afterSunset = false
	}

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
