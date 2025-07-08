package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func AfterNow(date, now time.Time) bool {
	return date.After(now)
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	date, err := time.Parse(Format, dstart)
	if err != nil {
		return "", errors.New("invalid format date")
	}

	rules := strings.Split(repeat, " ")
	switch rules[0] {
	case "d":
		if len(rules) < 2 || rules[1] == "" {
			return "", errors.New("interval of days isn't specified")
		}

		days, err := strconv.Atoi(rules[1])
		if err != nil {
			return "", errors.New("invalid format date in rule")
		}

		if days < 1 || days > 400 {
			return "", errors.New("days cannot be less than 400")
		}
		for {
			date = date.AddDate(0, 0, days)
			if AfterNow(date, now) {
				break
			}
		}

		return date.Format(Format), nil

	case "y":
		for {
			date = date.AddDate(1, 0, 0)
			if AfterNow(date, now) {
				break
			}
		}

		return date.Format(Format), nil

	case "w":
		if len(rules) < 2 || rules[1] == "" {
			return "", errors.New("days isn't specified")
		}
		daysStr := strings.Split(rules[1], ",")
		targetDays := make([]int, 0, len(daysStr))
		for _, d := range daysStr {
			day, err := strconv.Atoi(strings.TrimSpace(d))
			if err != nil || day < 1 || day > 7 {
				return "", errors.New("invalid day of week")
			}
			if day == 7 {
				targetDays = append(targetDays, 0)
			} else {
				targetDays = append(targetDays, day)
			}
		}

		start := now.Add(24 * time.Hour)
		if AfterNow(date, now) {
			start = date
		}
		currentWD := int(start.Weekday())

		minDays := 7

		for _, target := range targetDays {
			diff := (target - currentWD + 7) % 7
			if diff < minDays {
				minDays = diff
			}
		}

		result := start.AddDate(0, 0, minDays)
		return result.Format(Format), nil
	case "m":
		if len(rules) < 2 || rules[1] == "" {
			return "", errors.New("days not specified")
		}
		dayStr := strings.Split(rules[1], ",")
		days := make([]int, 0, len(dayStr))
		for _, d := range dayStr {
			day, err := strconv.Atoi(strings.TrimSpace(d))
			if err != nil || day < -2 || day > 31 || day == 0 {
				return "", errors.New("invalid day format")
			}
			days = append(days, day)
		}

		months := []int{}
		if len(rules) == 3 {
			monthStr := strings.Split(rules[2], ",")
			for _, m := range monthStr {
				month, err := strconv.Atoi(strings.TrimSpace(m))
				if err != nil || month < 1 || month > 12 {
					return "", errors.New("invalid month format")
				}
				months = append(months, month)
			}
		}

		start := now.Add(24 * time.Hour)
		if AfterNow(date, start) {
			start = date
		}
		startDate := start.Format(Format)
		candidateStr := []string{}
		currentYear := start.Year()

		for i := 0; i < 3; i++ {
			year := currentYear + i

			checkMonths := months
			if len(checkMonths) == 0 {
				checkMonths = make([]int, 12)
				for j := range checkMonths {
					checkMonths[j] = j + 1
				}
			}

			for _, month := range checkMonths {
				for _, day := range days {
					lastDay := time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, start.Location()).AddDate(0, 0, -1).Day()

					actualDay := day
					if day < 0 {
						actualDay = lastDay + day + 1
						if actualDay < 1 {
							continue
						}
					}

					if actualDay > lastDay {
						continue
					}

					candidate := fmt.Sprintf("%04d%02d%02d", year, month, actualDay)

					if _, err := time.Parse(Format, candidate); err == nil && candidate >= startDate {
						candidateStr = append(candidateStr, candidate)
					}
				}
			}
		}

		if len(candidateStr) == 0 {
			return "", errors.New("no valid date found")
		}

		result := candidateStr[0]
		for _, cand := range candidateStr[1:] {
			if cand < result {
				result = cand
			}
		}

		return result, nil

	default:
		return "", errors.New("invalid repeat")
	}
}
