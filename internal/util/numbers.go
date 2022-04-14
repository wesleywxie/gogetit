package util

import (
	"regexp"
	"strconv"
)

var (
	numberRegexp = regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
)

func ExtractInt(str string) (num int) {
	num = 0
	if numberRegexp.MatchString(str) {
		nums := numberRegexp.FindAllString(str, -1)
		if len(nums) > 0 {
			num, _ = strconv.Atoi(nums[0])
		}
	}
	return num
}

func ExtractFloat(str string) (num float64) {
	num = 0
	if numberRegexp.MatchString(str) {
		nums := numberRegexp.FindAllString(str, -1)
		if len(nums) > 0 {
			num, _ = strconv.ParseFloat(nums[0], 64)
		}
	}
	return num
}
