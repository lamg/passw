package main

import (
	"bytes"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"os/exec"
	"time"
)

type PsFile map[string][]Pass

type Pass struct {
	Date     time.Time `yaml:"date"`
	Password string    `yaml:"password"`
}

func main() {
	var res, fl string
	var erm, c, d, ls bool
	flag.StringVar(&res, "r", "", "Resource name")
	flag.BoolVar(&erm, "e", false,
		"Generate an easy to remind password")
	flag.BoolVar(&c, "c", false,
		"Create new password for resource")
	flag.BoolVar(&ls, "l", false, "List resources")
	flag.BoolVar(&d, "d", false, "Delete resource")
	flag.StringVar(&fl, "f", "", "Filename")
	flag.Parse()

	var pf PsFile
	var e error
	if fl != "" {
		pf = make(map[string][]Pass)
		_, e = os.Stat(fl)
		var bs []byte
		if e == nil {
			cmd := exec.Command("gpg", "--decrypt", fl)
			bs, e = cmd.Output()
		}
		if e == nil {
			e = yaml.Unmarshal(bs, pf)
		}
		if os.IsNotExist(e) {
			e = nil
		}
	} else {
		e = fmt.Errorf("No file specified")
	}
	if e == nil && c {
		// create resource
		_, ok := pf[res]
		if !ok {
			pf[res] = make([]Pass, 0)
		}
		p := Pass{Date: time.Now()}
		var cmd *exec.Cmd
		if erm {
			cmd = exec.Command("apg", "-n", "1")
		} else {
			cmd = exec.Command("apg", "-m", "16", "-a", "1", "-n", "1")
		}
		bs, e := cmd.Output()
		if e == nil {
			p.Password = string(bs)
			pf[res] = append(pf[res], p)
			bs, e = yaml.Marshal(pf)
		}
		if e == nil {
			// write to file
			cme := exec.Command("gpg", "--encrypt",
				"--default-recipient-self", "--armor", "--output", fl)
			cme.Stdin = bytes.NewBuffer(bs)
			e = cme.Run()
		}
	} else if e == nil && d {
		// delete resource
		delete(pf, res)
		// write to file
	} else if e == nil {
		// get resource
		p, ok := pf[res]
		var bs []byte
		if ok && !ls {
			// get last password of resource
			bs, e = yaml.Marshal(p[len(p)-1])
		} else if ok && ls {
			// list all passwords of resource
			bs, e = yaml.Marshal(p)
		} else {
			e = fmt.Errorf("%s not found", res)
		}
		if e == nil {
			fmt.Println(string(bs))
		}
	}
	ex := 0
	if e != nil {
		fmt.Fprintln(os.Stderr, e.Error())
		ex = 1
	}
	os.Exit(ex)
}
