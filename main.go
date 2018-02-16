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

type PsFile map[string]map[string][]Pass

type Pass struct {
	Date     time.Time `yaml:"date"`
	Password string    `yaml:"password"`
}

func main() {
	var res, fl, psw, usr string
	var erm, c, d, ls, g bool

	flag.StringVar(&fl, "f", "", "Filename")
	flag.StringVar(&res, "r", "", "Resource name")
	flag.StringVar(&usr, "u", "", "User name")

	flag.BoolVar(&c, "c", false,
		"Create new password for resource")
	flag.BoolVar(&g, "g", false, "Generate a new password")
	flag.BoolVar(&erm, "e", false,
		"Generate an easy to remind password")
	flag.StringVar(&psw, "p", "",
		"Predefined password for resource to be created")

	flag.BoolVar(&d, "d", false, "Delete resource")

	flag.BoolVar(&ls, "l", false, "List all passwords for resource")
	flag.Parse()

	pf, e := readFile(fl)
	if e == nil {
		if c || d {
			if c {
				e = crtRes(g, erm, pf, res, usr, psw)
			} else if d {
				e = delRes(pf, res)
			}
			if e == nil {
				e = writeToFl(pf, fl)
			}
		} else {
			e = getRes(pf, res, usr, ls)
		}
	}

	ex := 0
	if e != nil {
		fmt.Fprintln(os.Stderr, e.Error())
		ex = 1
	}
	os.Exit(ex)
}

func readFile(fl string) (pf PsFile, e error) {
	if fl != "" {
		pf = make(map[string]map[string][]Pass)
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
	return
}

func crtRes(g, erm bool, pf PsFile, res, usr, psw string) (e error) {
	_, ok := pf[res]
	if !ok {
		pf[res] = make(map[string][]Pass)
		pf[res][usr] = make([]Pass, 0)
	}
	p := Pass{Date: time.Now()}
	if g {
		var cmd *exec.Cmd
		if erm {
			cmd = exec.Command("apg", "-n", "1")
		} else {
			cmd = exec.Command("apg", "-m", "16", "-a", "1", "-n", "1")
		}
		var bs []byte
		bs, e = cmd.Output()
		if e == nil {
			psw = string(bs)
		}
	}
	if psw == "" {
		e = fmt.Errorf("Empty string cannot be used as password")
	}
	if e == nil {
		p.Password = psw
		_, ok = pf[res][usr]
		if !ok {
			pf[res][usr] = make([]Pass, 0)
		}
		pf[res][usr] = append(pf[res][usr], p)
	}
	return
}

func delRes(pf PsFile, res string) (e error) {
	_, ok := pf[res]
	if ok {
		delete(pf, res)
	} else {
		e = fmt.Errorf("Resource %s doesn't exists", res)
	}
	return
}

func writeToFl(pf PsFile, fl string) (e error) {
	var bs []byte
	bs, e = yaml.Marshal(pf)
	if e == nil {
		cme := exec.Command("gpg", "--encrypt",
			"--default-recipient-self", "--armor", "--output", fl)
		cme.Stdin = bytes.NewBuffer(bs)
		e = cme.Run()
	}
	return
}

func getRes(pf PsFile, res, usr string, ls bool) (e error) {
	acs, ok := pf[res]
	if !ok {
		e = fmt.Errorf("%s not found", res)
	}
	var bs []byte
	if e == nil {
		if usr == "" {
			bs, e = currPsAll(acs)
		} else {
			bs, e = currPsUsr(acs, usr)
		}
	}
	if e == nil {
		fmt.Println(string(bs))
	}
	return
}

type CurrPs struct {
	User string `yaml:"user"`
	Pass Pass   `yaml:"pass"`
}

func currPsAll(acs map[string][]Pass) (bs []byte, e error) {
	rs, i := make([]CurrPs, len(acs)), 0
	for k, v := range acs {
		rs[i], i = CurrPs{User: k, Pass: v[len(v)-1]}, i+1
	}
	bs, e = yaml.Marshal(rs)
	return
}

func currPsUsr(acs map[string][]Pass, usr string) (bs []byte,
	e error) {
	sp, ok := acs[usr]
	if ok {
		p := sp[len(sp)-1]
		bs, e = yaml.Marshal(p)
	} else {
		e = fmt.Errorf("Not found user %s", usr)
	}
	return
}
