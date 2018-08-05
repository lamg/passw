package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

// PsFile is the data structure being [de]serialized in the text
// file using YAML format
type PsFile map[string]map[string][]Pass

// Pass is the data structure representing a created password
type Pass struct {
	// Date when password was created
	Date     time.Time `yaml:"date"`
	Password string    `yaml:"password"`
}

func main() {
	var r, u, f string
	var c bool

	flag.StringVar(&f, "f", "", "File name")
	flag.StringVar(&r, "r", "", "Resource name")
	flag.StringVar(&u, "u", "", "User name")
	flag.BoolVar(&c, "c", false, "Create new password for resource")

	var easy bool
	flag.BoolVar(&easy, "e", false, "Call 'apg -n 1' or 'apg -n 1 -a 1 -m 16'")

	flag.Parse()

	af := &afero.Afero{Fs: afero.NewOsFs()}
	pf, e := readFile(f, af)
	var password string
	if !c {
		// first use case call
		password, e = retrieve(r, u, pf)
	} else {
		// second use case call
		password, e = create(f, r, u, pf, easy, af)
	}
	ex := 0
	if e == nil {
		fmt.Println(password)
	} else {
		fmt.Fprint(os.Stderr, e.Error())
		ex = 1
	}
	os.Exit(ex)
}

func readFile(fl string, af *afero.Afero) (pf PsFile, e error) {
	var bs []byte
	bs, e = af.ReadFile(fl)
	if e == nil {
		pf = make(map[string]map[string][]Pass)
		e = yaml.Unmarshal(bs, pf)
	}
	return
}

func retrieve(resource, user string, pf PsFile) (r string, e error) {
	var ps []Pass
	ps, e = retrPs(resource, user, pf)
	if e == nil {
		r = ps[len(ps)-1].Password
	}
	return
}

func noUser(user string) (e error) {
	e = fmt.Errorf("No user %s", user)
	return
}

func noResource(resource string) (e error) {
	e = fmt.Errorf("No resource %s", resource)
	return
}

func create(file, resource, user string, pf PsFile, easy bool, af *afero.Afero) (r string, e error) {
	var cmd *exec.Cmd
	if easy {
		cmd = exec.Command("apg", "-n", "1")
	} else {
		cmd = exec.Command("apg", "-m", "16", "-a", "1", "-n", "1")
	}
	var bs []byte
	bs, e = cmd.Output()
	if e == nil {
		// deleting new line character
		r = strings.TrimRight(string(bs), "\n\r")
		_, e = retrPs(resource, user, pf)
	}
	var sr []byte
	if e == nil {
		pf[resource][user] = append(pf[resource][user], Pass{
			Date:     time.Now(),
			Password: r,
		})
		sr, e = yaml.Marshal(pf)
	}
	if e == nil {
		e = af.WriteFile(file, sr, os.ModePerm)
	}
	return
}

func retrPs(resource, user string, pf PsFile) (r []Pass, e error) {
	um, ok := pf[resource]
	if ok {
		r, ok = um[user]
		if !ok || len(r) == 0 {
			e = noUser(user)
		}
	} else {
		e = noResource(resource)
	}
	return
}
