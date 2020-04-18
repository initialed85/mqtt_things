package circumstances_engine

import (
	"fmt"
	"log"
	"strings"
	"time"
)

type Circumstances struct {
	Timestamp                                              time.Time
	AfterSunrise, AfterSunset, AfterBedtime, AfterWaketime bool
	Hot, Comfortable, Cold                                 bool
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
		afterSunrise, afterSunset,
		afterBedtime, afterWaketime,
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

	sunrise = recreateForNow(sunrise.Add(offset), now)
	sunset = recreateForNow(sunset.Add(offset), now)
	bedtime = recreateForNow(bedtime.Add(offset), now)
	waketime = recreateForNow(waketime.Add(offset), now)

	log.Printf("recreated sunrise is %v", sunrise)
	log.Printf("recreated sunset is %v", sunset)
	log.Printf("recreated bedtime is %v", bedtime)
	log.Printf("recreated waketime is %v", waketime)

	if now.After(sunset) {
		afterSunrise = false
		afterSunset = true
	} else if now.After(sunrise) {
		afterSunrise = true
		afterSunset = false
	} else {
		afterSunrise = false
		afterSunset = true
	}

	if now.After(bedtime) {
		afterWaketime = false
		afterBedtime = true
	} else if now.After(waketime) {
		afterWaketime = true
		afterBedtime = false
	} else {
		afterWaketime = false
		afterBedtime = true
	}

	cold = temperature <= coldEntry
	comfortable = temperature >= coldExit && temperature <= hotExit
	hot = temperature >= hotEntry

	circumstances := Circumstances{
		Timestamp:     now,
		AfterSunrise:  afterSunrise,
		AfterSunset:   afterSunset,
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
		{fmt.Sprintf("%v/after_sunrise%v/get", prefix, suffix), convertCircumstance(circumstances.AfterSunrise)},
		{fmt.Sprintf("%v/after_sunset%v/get", prefix, suffix), convertCircumstance(circumstances.AfterSunset)},
		{fmt.Sprintf("%v/after_bedtime%v/get", prefix, suffix), convertCircumstance(circumstances.AfterBedtime)},
		{fmt.Sprintf("%v/after_waketime%v/get", prefix, suffix), convertCircumstance(circumstances.AfterWaketime)},
		{fmt.Sprintf("%v/hot%v/get", prefix, suffix), convertCircumstance(circumstances.Hot)},
		{fmt.Sprintf("%v/comfortable%v/get", prefix, suffix), convertCircumstance(circumstances.Comfortable)},
		{fmt.Sprintf("%v/cold%v/get", prefix, suffix), convertCircumstance(circumstances.Cold)},
	}
}
