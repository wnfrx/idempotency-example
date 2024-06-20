package util

import "strconv"

func StringToInt(value string, def ...int) (result int) {
	var err error
	result, err = strconv.Atoi(value)
	if err != nil {
		if len(def) > 0 {
			return def[0]
		}
		return 0
	}
	return result
}
