package cron

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type CronExpression struct {
	Year int
	Month int
	Day int
	Hour int
	Minute int
	Second int
	IsEnd bool
	MoveNextYear func() time.Time
	MoveNextMonth func() time.Time
	MoveNextDay func() time.Time
	MoveNextHour func() time.Time
	MoveNextMinute func() time.Time
	MoveNextSecond func() time.Time
	CheckYear func() bool
	CheckMonth func() bool
	CheckDay func() bool
	CheckHour func() bool
	CheckMinute func() bool
	CheckSecond func() bool
}

func ParseCronExpression(line string) *CronExpression {
	regexLine := regexp.MustCompile(`^(?P<second>(.*?))\s+(?P<minute>(.*?))\s+(?P<hour>(.*?))\s+(?P<dayofmonth>(.*?))\s+(?P<month>(.*?))\s+(?P<dayofweek>(.*?))\s+(?P<year>(.*?))$`)
	match := regexLine.FindStringSubmatch(line)
	if match == nil {
		panic(line)
	}

	result := make(map[string]string)
	groupNames := regexLine.SubexpNames()
	for i, name := range groupNames {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	if result["dayofmonth"] != "?" && result["dayofweek"] != "?" {
		panic(line)
	}

	now := time.Now()
	ce := &CronExpression{Second:now.Second(),Minute:now.Minute(),Hour:now.Hour(),Day:now.Day(),Month:int(now.Month()),Year:now.Year(),IsEnd:false}
	for k, v := range result {
		flag := false
		for _, r := range _cronPatternCheck[k] {
			if r.MatchString(v) {
				if _cronValueCheck[r](k, v) == false {
					panic(v)
				}
				if v != "?" {
					switch k {
					case "second":
						ce.MoveNextSecond, ce.CheckSecond = _cronFunc[r](ce, k, v)
					case "minute":
						ce.MoveNextMinute, ce.CheckMinute = _cronFunc[r](ce, k, v)
					case "hour":
						ce.MoveNextHour, ce.CheckHour = _cronFunc[r](ce, k, v)
					case "dayofweek", "dayofmonth":
						ce.MoveNextDay, ce.CheckDay = _cronFunc[r](ce, k, v)
					case "month":
						ce.MoveNextMonth, ce.CheckMonth = _cronFunc[r](ce, k, v)
					case "year":
						ce.MoveNextYear, ce.CheckYear = _cronFunc[r](ce, k, v)
					}
				}
				flag = true
				break
			}
		}
		if !flag {
			panic(fmt.Sprintf("expression:%s, unknown:%s", line, v))
		}
	}
	return ce
}

func (ce *CronExpression) MoveNext() time.Time {
	if ce.IsEnd {
		return ce.ToTime()
	}
	now := ce.ToTime()
	if !ce.CheckSecond() || ce.ToTime().Unix() <= now.Unix() {
		ce.MoveNextSecond()
	}
	if !ce.CheckMinute() || ce.ToTime().Unix() <= now.Unix() {
		ce.MoveNextMinute()
	}
	if !ce.CheckHour() || ce.ToTime().Unix() <= now.Unix() {
		ce.MoveNextHour()
	}
	if !ce.CheckDay() || ce.ToTime().Unix() <= now.Unix() {
		ce.MoveNextDay()
	}
	if !ce.CheckMonth() || ce.ToTime().Unix() <= now.Unix() {
		ce.MoveNextMonth()
	}
	if !ce.CheckYear() || ce.ToTime().Unix() <= now.Unix() {
		ce.MoveNextYear()
	}
	if ce.ToTime().Unix() <= now.Unix() {
		ce.IsEnd = true
		ce.SetTime(now)
	}
	return ce.ToTime()
}

func (ce *CronExpression) GetValue(timePart string) int {
	switch timePart {
	case "second":
		return ce.Second
	case "minute":
		return ce.Minute
	case "hour":
		return ce.Hour
	case "dayofweek","dayofmonth","day":
		return ce.Day
	case "month":
		return ce.Month
	case "year":
		return ce.Year
	}
	panic(timePart)
}

func (ce *CronExpression) SetValue(timePart string, val int) *CronExpression {
	switch timePart {
	case "second":
		ce.Second = val
	case "minute":
		ce.Minute = val
		ce.SetValue("second",0)
		if !ce.CheckMinute() {
			ce.MoveNextSecond()
		}
	case "hour":
		ce.Hour = val
		ce.SetValue("minute",0)
		if !ce.CheckMinute() {
			ce.MoveNextMinute()
		}
	case "dayofweek","dayofmonth","day":
		ce.Day = val
		ce.SetValue("hour",0)
		if !ce.CheckHour() {
			ce.MoveNextHour()
		}
	case "month":
		ce.Month = val
		ce.SetValue("day",1)
		if !ce.CheckDay() {
			ce.MoveNextDay()
		}
		if !ce.CheckDay() {
			ce.MoveNextMonth()
		}
	case "year":
		ce.Year = val
		ce.SetValue("month",1)
		if !ce.CheckMonth() {
			ce.MoveNextMonth()
		}
	}
	return ce
}

func (ce *CronExpression) SetTime(t time.Time) *CronExpression {
	ce.Year = t.Year()
	ce.Month = int(t.Month())
	ce.Day = t.Day()
	ce.Hour = t.Hour()
	ce.Minute = t.Minute()
	ce.Second = t.Second()
	return ce
}

func (ce *CronExpression) ToTime() time.Time {
	return time.Date(ce.Year,time.Month(ce.Month),ce.Day,ce.Hour,ce.Minute,ce.Second,0,time.Local)
}

var _regexStar = regexp.MustCompile(`^\*$`) //eg: *
var _regexArea = regexp.MustCompile(`^([0-9]+)-([0-9]+)$`) //eg: 20-30
var _regexSlice = regexp.MustCompile(`^(([0-9]+)|\*)/[0-9]+$`) //eg: */10
var _regexAreaSlice = regexp.MustCompile(`^([0-9]+)-([0-9]+)/[0-9]+$`) //eg: 10-30/2
var _regexEnum = regexp.MustCompile(`^([0-9]+)(,[0-9]+)*$`) //eg: 10,20,30,40
var _regexL = regexp.MustCompile(`^[1-7]?L$`) //eg: 4L
var _regexLOnly = regexp.MustCompile(`^L$`)
var _regexLW = regexp.MustCompile(`^LW$`) //eg: LW
var _regexIgnore = regexp.MustCompile(`^\?$`) //eg: ?
var _regexWeekDay = regexp.MustCompile(`^[1-7]#[1-5]$`) //eg:  3#2

var _cronPatternCheck = map[string][]*regexp.Regexp {
	"second": []*regexp.Regexp { _regexStar, _regexArea, _regexSlice, _regexAreaSlice, _regexEnum},
	"minute": []*regexp.Regexp { _regexStar, _regexArea, _regexSlice, _regexAreaSlice, _regexEnum},
	"hour": []*regexp.Regexp { _regexStar, _regexArea, _regexSlice, _regexAreaSlice, _regexEnum},
	"dayofmonth": []*regexp.Regexp { _regexStar, _regexArea, _regexSlice, _regexAreaSlice, _regexEnum, _regexLW , _regexLOnly, _regexIgnore},
	"month": []*regexp.Regexp { _regexStar, _regexArea, _regexSlice, _regexAreaSlice, _regexEnum},
	"dayofweek": []*regexp.Regexp { _regexStar, _regexArea, _regexSlice, _regexAreaSlice, _regexEnum, _regexL, _regexIgnore, _regexWeekDay},
	"year": []*regexp.Regexp { _regexStar, _regexArea, _regexSlice, _regexAreaSlice, _regexEnum},
}

var _timeRange = map[string][]int {
	"second": []int { 0, 59},
	"minute": []int { 0, 59},
	"hour": []int { 0, 23},
	"dayofmonth": []int { 1, 31},
	"month": []int { 1, 12},
	"dayofweek": []int { 1, 7},
	"year": []int { time.Now().Year(), time.Now().Year() + 100},
}

var _cronValueCheck = map[*regexp.Regexp]func(timePart string, match string) bool {
	_regexStar: func(timePart string, match string) bool { return true },
	_regexArea: func(timePart string, match string) bool {
		areas := strings.Split(match, "-")
		start, _ := strconv.Atoi(areas[0])
		end, _ := strconv.Atoi(areas[1])
		min := _timeRange[timePart][0]
		max := _timeRange[timePart][1]
		return !(start < min || start > max || end < min || end > max)
	},
	_regexSlice: func(timePart string, match string) bool {
		parts := strings.Split(match, "/")
		start := parts[0]
		slice, _  := strconv.Atoi(parts[1])
		min := _timeRange[timePart][0]
		max := _timeRange[timePart][1]
		if start == "*" {
			return !(slice < min || slice > max)
		} else {
			iStart, _ := strconv.Atoi(start)
			return !(iStart < min || iStart > max || slice < min || slice > max)
		}
	},
	_regexAreaSlice: func(timePart string, match string) bool {
		parts := strings.Split(match, "/")
		areas := strings.Split(parts[0], "-")
		start, _ := strconv.Atoi(areas[0])
		end, _ := strconv.Atoi(areas[1])
		slice, _  := strconv.Atoi(parts[1])
		min := _timeRange[timePart][0]
		max := _timeRange[timePart][1]
		return !(start < min || start > max || end < min || end > max || slice < min || slice > max)
	},
	_regexEnum: func(timePart string, match string) bool {
		min := _timeRange[timePart][0]
		max := _timeRange[timePart][1]
		for _, enum := range strings.Split(match, ",") {
			num, _ := strconv.Atoi(enum)
			if num < min || num > max {
				return false
			}
		}
		return true
	},
	_regexL: func(timePart string, match string) bool {
		parts := strings.Split(match, "L")
		num, err := strconv.Atoi(parts[0])
		if err != nil {
			return true
		}
		min := _timeRange[timePart][0]
		max := _timeRange[timePart][1]
		return !(num < min || num > max)
	},
	_regexLOnly: func(timePart string, match string) bool {
		return true
	},
	_regexLW: func(timePart string, match string) bool {
		return true
	},
	_regexIgnore: func(timePart string, match string) bool {
		return true
	},
	_regexWeekDay: func(timePart string, match string) bool {
		parts := strings.Split(match, "#")
		day, _ := strconv.Atoi(parts[0])
		week, _ := strconv.Atoi(parts[1])
		return 1 <= day && day <= 7 && 1 <= week && week <= 5
	},
}

var _cronFunc = map[*regexp.Regexp]func(ce *CronExpression, timePart string, match string) (func() time.Time, func() bool) {
	_regexStar: func(ce *CronExpression, timePart string, match string) (func() time.Time, func() bool) {
		min := _timeRange[timePart][0]
		max := _timeRange[timePart][1]
		slices := Slice(min, max, min, max, 1)
		sort.Ints(slices)
		return CreateMoveFunc(ce, timePart, slices),func() bool { return true }
	},
	_regexArea: func(ce *CronExpression, timePart string, match string) (func() time.Time, func() bool) {
		areas := strings.Split(match, "-")
		start, _ := strconv.Atoi(areas[0])
		end, _ := strconv.Atoi(areas[1])
		min := _timeRange[timePart][0]
		max := _timeRange[timePart][1]
		slices := Slice(start, end, min, max, 1)
		sort.Ints(slices)
		return CreateMoveFunc(ce, timePart, slices), CreateCheckFunc(ce,timePart,slices)
	},
	_regexSlice: func(ce *CronExpression, timePart string, match string) (func() time.Time, func() bool) {
		parts := strings.Split(match, "/")
		start := parts[0]
		slice, _  := strconv.Atoi(parts[1])
		min := _timeRange[timePart][0]
		max := _timeRange[timePart][1]
		slices := []int{}
		if start == "*" {
			slices = Slice(min, max, min, max, slice)
		} else {
			iStart, _ := strconv.Atoi(start)
			slices = Slice(iStart, max, min, max, slice)
		}
		sort.Ints(slices)
		return CreateMoveFunc(ce, timePart, slices), CreateCheckFunc(ce,timePart,slices)
	},
	_regexAreaSlice: func(ce *CronExpression, timePart string, match string) (func() time.Time, func() bool) {
		parts := strings.Split(match, "/")
		areas := strings.Split(parts[0], "-")
		start, _ := strconv.Atoi(areas[0])
		end, _ := strconv.Atoi(areas[1])
		slice, _  := strconv.Atoi(parts[1])
		min := _timeRange[timePart][0]
		max := _timeRange[timePart][1]
		slices := Slice(start, end, min, max, slice)
		sort.Ints(slices)
		return CreateMoveFunc(ce, timePart, slices), CreateCheckFunc(ce,timePart,slices)
	},
	_regexEnum: func(ce *CronExpression, timePart string, match string) (func() time.Time, func() bool) {
		slices := []int{}
		for _, s := range strings.Split(match, ",") {
			num, _ := strconv.Atoi(s)
			slices = append(slices, num)
		}
		sort.Ints(slices)
		return CreateMoveFunc(ce, timePart, slices), CreateCheckFunc(ce,timePart,slices)
	},
	_regexLOnly: func(ce *CronExpression, timePart string, match string) (func() time.Time, func() bool) {
		return func() time.Time {
			if ce.CheckDay() {
				return ce.ToTime()
			}
			nextDate := time.Time{}
			nextMonth :=  time.Date(ce.Year,time.Month(ce.Month),1,0,0,0,0, time.Local).AddDate(0,1,0)
			nextDate = nextMonth.AddDate(0,0,-1)
			ce.SetValue(timePart, nextDate.Day())
			return ce.ToTime()
		}, func() bool {
				return ce.ToTime().AddDate(0,0,1).Month() != time.Month(ce.Month)
			}
	},
	_regexL: func(ce *CronExpression, timePart string, match string) (func() time.Time, func() bool) {
		return func() time.Time {
			now := ce.ToTime()
			nextDate := time.Time{}
			parts := strings.Split(match, "L")
			if parts[0] == "" { //each weekend
				tmp := time.Date(now.Year(), now.Month(), now.Day(), 0 ,0,0,0, time.Local)
				for  {
					tmp := tmp.AddDate(0 , 0, 1)
					if tmp.Month() != now.Month() {
						return ce.ToTime()
					}
					if tmp.Weekday() != time.Saturday {
						continue
					}
					nextDate = tmp
					break
				}
			} else {
				want, _ := strconv.Atoi(parts[0])
				nextDate = now
				nextMonth :=  time.Date(ce.Year,time.Month(ce.Month),1,0,0,0,0, time.Local).AddDate(0,1,0)
				for i := 1; i <= 7 ; i++ {
					tmp := nextMonth.AddDate(0 , 0, -i)
					if int(tmp.Weekday()) + 1 != want  {
						continue
					}
					nextDate = tmp
					break
				}
			}
			ce.SetValue(timePart, nextDate.Day())
			return ce.ToTime()
		}, func() bool {
				now := ce.ToTime()
				parts := strings.Split(match, "L")
				if parts[0] == "" {
					return now.Weekday() == time.Saturday
				}else{
					want, _ := strconv.Atoi(parts[0])
					return int(now.Weekday()) + 1 == want && now.AddDate(0,0,7).Month() != time.Month(ce.Month)
				}
			}
	},
	_regexLW: func(ce *CronExpression, timePart string, match string) (func() time.Time, func() bool) {
		return func() time.Time {
			nextDate := time.Time{}
			nextMonth :=  time.Date(ce.Year,time.Month(ce.Month),1,0,0,0,0, time.Local).AddDate(0,1,0)
			for i := 1; i <= 7 ; i++ {
				tmp := nextMonth.AddDate(0 , 0, -i)
				if tmp.Weekday() == time.Saturday || tmp.Weekday() == time.Sunday {
					continue
				}
				nextDate = tmp
				break
			}
			ce.SetValue(timePart, nextDate.Day())
			return ce.ToTime()
		}, func() bool {
				nextMonth :=  time.Date(ce.Year,time.Month(ce.Month),1,0,0,0,0, time.Local).AddDate(0,1,0)
				tmp := time.Time{}
				for i := 1; i <= 7 ; i++ {
					tmp = nextMonth.AddDate(0 , 0, -i)
					if tmp.Weekday() == time.Saturday || tmp.Weekday() == time.Sunday {
						continue
					}
					break;
				}
				return ce.Day == tmp.Day()
			}
	},
	_regexWeekDay: func(ce *CronExpression, timePart string, match string) (func() time.Time, func() bool) {
		parts := strings.Split(match, "#")
		day, _ := strconv.Atoi(parts[0])
		num, _ := strconv.Atoi(parts[1])
		return func() time.Time {
			now := ce.ToTime()
			start := time.Date(now.Year(),now.Month(),1,0,0,0,0, time.Local)
			count := 0
			for i:=0; i<31; i++{
				if int(start.Weekday()) + 1 == day {
					count++
					if count == num {
						break
					}
				}
				start = start.AddDate(0,0,1)
			}
			if start.Month() != now.Month() {
				return ce.ToTime()
			} else {
				ce.SetValue(timePart, start.Day())
				return ce.ToTime()
			}
		}, func() bool {
				now := ce.ToTime()
				start := time.Date(now.Year(),now.Month(),1,0,0,0,0, time.Local)
				count := 0
				for i:=0; i<31; i++{
					if int(start.Weekday()) + 1 == day {
						count++
						if count == num {
							break
						}
					}
					start = start.AddDate(0,0,1)
				}
				return start.Month() == now.Month() && start.Day() == now.Day()
			}
	},
}

var _valMap = map[string]func(ce *CronExpression, vals []int) []int {
	"dayofweek": func(ce *CronExpression, vals []int) []int {
		d := make(map[int]bool)
		for _, val := range vals {
			d[val] = true
		}
		days := []int{}
		start := time.Date(ce.Year,time.Month(ce.Month),ce.Day,0,0,0,0, time.Local)
		for start.Month() == time.Month(ce.Month) {
			if d[int(start.Weekday())+1] {
				days = append(days, start.Day())
			}
			start = start.AddDate(0,0,1)
		}
		return days
	},
	"dayofmonth": func(ce *CronExpression, vals []int) []int {
		days := []int{}
		for _, val := range vals {
			if time.Date(ce.Year,time.Month(ce.Month),val,0,0,0,0,time.Local).Month() == time.Month(ce.Month) {
				days = append(days, val)
			} else {
				break
			}
		}
		return days
	},
	"year": func(ce *CronExpression, vals []int) []int {
		checkYear := func(year int) bool {
			if ce.Month == 2 && ce.Day == 29 {
				return year % 4 == 0 && year % 100 != 0
			} else {
				return true
			}
		}
		years := []int{}
		for _, val := range vals {
			if checkYear(val) {
				years = append(years, val)
			}
		}
		return years
	},
}

func NextValue(ce *CronExpression, timePart string, val int, nums []int) int {
	if len(nums) == 0 {
		return -1
	}
	for i := 1; i < len(nums); i++ {
		if val >= nums[i-1] && val < nums[i] {
			return nums[i]
		}
	}
	return nums[0]
}

func CreateMoveFunc(ce *CronExpression, timePart string, slices []int) func() time.Time {
	return func() time.Time {
		val := ce.GetValue(timePart)
		vals := slices
		if f, exists := _valMap[timePart]; exists {
			vals = f(ce, vals)
		}
		nextVal := NextValue(ce , timePart, val, vals)
		if nextVal == -1 {
			return ce.ToTime()
		}
		ce.SetValue(timePart, nextVal)
		return ce.ToTime()
	}
}

func CreateCheckFunc(ce *CronExpression, timePart string, slices []int) func() bool {
	return func() bool {
		vals := slices
		if f, exists := _valMap[timePart]; exists {
			vals = f(ce, vals)
		}
		val := ce.GetValue(timePart)
		for _, num := range vals {
			if num == val {
				return true
			}
		}
		return false
	}
}

func Slice(start, end , min , max, slice int) []int {
	if slice == 0 {
		return []int {start}
	}
	set := []int{}
	if start > end {
		for i := start ; i <= max ; i += slice {
			set = append(set, i)
		}
		for i := min ; i <= end ; i += slice {
			set = append(set, i)
		}
	} else {
		for i := start ; i <= end ; i += slice {
			set = append(set, i)
		}
	}
	nums := []int{}
	for i := 0; i < len(set); i ++ {
		nums = append(nums, set[i])
	}
	return nums
}