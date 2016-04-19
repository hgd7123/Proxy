package main

import (
	"strings"
)

func ConvertStr2Map(in string) (out map[string]string) {
	outm := make(map[string]string)
	in = strings.Replace(in, "\r", "", -1)
	inarr := strings.Split(in, "\n")
	for _, v := range inarr {
		spa := strings.Split(v, ": ")
		if len(spa) > 1 {
			outm[spa[0]] = spa[1]
			//	fmt.Println(spa[0], spa[1])
		}
	}
	return outm
}
