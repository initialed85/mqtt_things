package circumstances_engine

import (
	"fmt"
	"log"
	"strings"
	"time"
)

type Circumstances struct {
	Timestamp                                                                                                          time.Time
	BeforeSunrise, AfterSunrise, BeforeSunset, AfterSunset, BeforeBedtime, AfterBedtime, BeforeWaketime, AfterWaketime bool
	Hot, Comfortable, Cold                                                                                             bool
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

func CalculateCircumstances(
	now, sunrise, sunset, bedtime, waketime time.Time,
	temperature, hotEntry, hotExit, coldEntry, coldExit float64,
	offset time.Duration,
) Circumstances {
	var (
		beforeSunrise, afterSunrise, beforeSunset, afterSunset,
		beforeBedtime, afterBedtime, beforeWaketime, afterWaketime,
		hot, comfortable, cold bool
	)

	log.Printf(
		"getting circumstances for %v, %v, %v, %v, %v, %v, %v, %v, %v, %v, %v",
		now,
		sunrise,
		sunset,
		bedtime,
		waketime,
		temperature,
		hotEntry,
		hotExit,
		coldEntry,
		coldExit,
		offset,
	)

	log.Printf("now is %v", now)

	sunrise = recreateForNow(sunrise, now)
	sunset = recreateForNow(sunset, now)
	bedtime = recreateForNow(bedtime, now)
	waketime = recreateForNow(waketime, now)

	log.Printf("sunrise is %v", sunrise)
	log.Printf("sunset is %v", sunset)
	log.Printf("bedtime is %v", bedtime)

	if now.Before(sunrise.Add(-offset)) || now.After(sunset.Add(-offset)) {
		beforeSunrise = true
	}

	if now.Before(waketime.Add(-offset)) || now.After(bedtime.Add(-offset)) {
		beforeWaketime = true
	}

	beforeSunset = !beforeSunrise
	beforeBedtime = !beforeWaketime

	afterSunrise = !beforeSunrise
	afterSunset = !beforeSunset
	afterWaketime = !beforeWaketime
	afterBedtime = !beforeBedtime

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
		AfterWaketime: afterWaketime,
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
		{fmt.Sprintf("%v/before_waketime%v/get", prefix, suffix), convertCircumstance(circumstances.BeforeWaketime)},
		{fmt.Sprintf("%v/after_waketime%v/get", prefix, suffix), convertCircumstance(circumstances.AfterWaketime)},
		{fmt.Sprintf("%v/hot%v/get", prefix, suffix), convertCircumstance(circumstances.Hot)},
		{fmt.Sprintf("%v/comfortable%v/get", prefix, suffix), convertCircumstance(circumstances.Comfortable)},
		{fmt.Sprintf("%v/cold%v/get", prefix, suffix), convertCircumstance(circumstances.Cold)},
	}
}
