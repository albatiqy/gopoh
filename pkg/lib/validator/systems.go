package validator

import (
	"regexp"
	"sort"
	"strconv"
	"time"
)

func NIK(input string) bool {
	if match, _ := regexp.MatchString("^[0-9]{16}$", input); !match {
		return false
	}

	var tgLhrStr string

	if tgLhr, err := strconv.Atoi(input[6:8]); err != nil {
		return false
	} else {
		if tgLhr > 40 {
			tgLhrStr = strconv.Itoa(tgLhr - 40)
		}
	}

	if tgLhrStr == "" {
		if _, err := time.Parse("010206", input[6:12]); err != nil {
			return false
		}
	} else {
		if _, err := time.Parse("010206", tgLhrStr+input[8:12]); err != nil {
			return false
		}
	}

	return true
}

func NIP(input string) bool {
	if match, _ := regexp.MatchString("^[0-9]{18}$", input); !match {
		return false
	}

	if _, err := time.Parse("20060102", input[:7]); err != nil {
		return false
	}
	if _, err := time.Parse("20060102", input[8:14]+"01"); err != nil {
		return false
	}
	if sort.SearchStrings([]string{"1", "2"}, input[14:15]) < 0 {
		return false
	}

	return true
}

func NPWP(input string) bool {
	if match, _ := regexp.MatchString("^[0-9]{15}$", input); !match {
		return false
	}

	return true
}

func DapodikKodeWilayah(input string) bool {
	if match, _ := regexp.MatchString("^[0-9 ]{6,8}$", input); !match {
		return false
	}

	return true
}

func DapodikNPSN(input string) bool {
	if match, _ := regexp.MatchString("^[0-9]{8}$", input); !match {
		return false
	}

	return true
}

func DapodikNUPTK(input string) bool {
	if match, _ := regexp.MatchString("^[0-9]{16}$", input); !match {
		return false
	}

	return true
}

func NomorSeluler(input string) bool {
	if match, _ := regexp.MatchString("^(08)[0-9]{8,20}", input); !match {
		return false
	}

	return true
}
