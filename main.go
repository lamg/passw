package main

import (
	"flag"
	"time"
)

type PsFile []rPass

type RPass struct {
	Res string `json:"res"`
	Ps  []Pass `json:"ps"`
}

type Pass struct {
	Date     time.Time `json:"date"`
	Password string    `json:"password"`
}

func main() {
	var res, fl string
	var prn, rw, ls bool
	flag.StringVar(&res, "r", "", "Resource name")
	flag.BoolVar(&prn, "e", false,
		"Generate an easy to remind password")
	flag.BoolVar(&rw, "c", false,
		"Create new password for resource")
	flag.BoolVar(&ls, "l", "false", "List resources")
	flag.BoolVar(&d, "d", false, "Delete resource")
	flag.StringVar(fl, "f", "", "Filename")
}
