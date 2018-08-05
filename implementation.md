# Password manager (pmng)

The pmng password manager is a simple program for storing and creating passwords (using `apg` http://apg.com) associated to resources and user names in a text file using YAML format. Created passwords are added to the supplied resource with the current date in RFC3339 format. It has two use cases:

## Retrieve
Retrieve a password given a file, resource and user name. A resource can be the web site name which has the account. User and password have the usual meaning. In this case the command(having F as the file where information is stored, R as resource name, and U as user name) is:

```sh
pmng -r R -u U -f F
```

and it outputs the retrieved password to standard output.

## Create
Create a password given a file, resource, user name, password and an `-e` flag which alternates between calling `apg -n 1` when present (this means it will generate a pronounceable password with length 8) and calling `apg -n 1 -a 1 -m 16`, which will generate a random character password with length 16 when absent. With R, U, F having the same above meaning and the `-c` flag distinguishing this call from the previous, the command is:

```sh
pmng -c [-e] -r R -u U -f F
```

and it outputs the created password to standard output.

## Implementation

The YAML format is selected for storing passwords in file because it is easy to read with standard tools. This means the labor can be focused in the two previous use cases without loosing usability, since all needed for seeing all passwords in the file or of a particular resource in the past is a text editor or a program like `cat` or `less`, which allow to use other tools for filtering text lines.

Before implementing use cases its needed implementing this data structure which is used for retrieving and storing passwords, and also the code selecting between the two use cases.

### Common data structure

The data structure is the following:

```go "data structures"
// PsFile is the data structure being [de]serialized in the text
// file using YAML format
type PsFile map[string]map[string][]Pass
```

which depends on

```go "data structures" +=
// Pass is the data structure representing a created password
type Pass struct {
  // Date when password was created
	Date     time.Time `yaml:"date"`
	Password string    `yaml:"password"`
}
```

and imports

```go "imports"
"time"
```

### Common flags

The common flags between both use cases are `-f`, `-u`, `-r`, and `-c`, which when absent means the first use case is selected, otherwise is the second. Parsing these flags is done by

```go "flags definition"
var r, u, f string
var c bool

flag.StringVar(&f, "f", "", "File name")
flag.StringVar(&r, "r", "", "Resource name")
flag.StringVar(&u, "u", "", "User name")
flag.BoolVar(&c, "c", false, "Create new password for resource")

```

and imports

```go "imports" +=
"flags"
```

Notice that `flag.Parse()` doesn't appear since only common flags are defined in this section, and could be more.

### Common procedures

Another piece of common code is the procedure for reading (deserializing) a file into the `PsFile` structure:

```go "common procedures"
func readFile(fl string, af *afero.Afero) (pf PsFile, e error) {
  var bs []byte
	bs, e = af.ReadFile(fl)
	if e == nil {
		pf = make(map[string]map[string][]Pass)
		e = yaml.Unmarshal(bs, pf)
  }
  return
}
```

which imports

```go "imports" +=
"github.com/spf13/afero"
"gopkg.in/yaml.v2"
```


### General code layout

Now with the common code implemented is possible to devise the code layout for the main procedure, having sections: _imports_, _data structures_, _common procedures_, _first use case procedures_ and _second use case procedures_.

```go main.go
import (
  <<<imports>>>
)

<<<data structures>>>

func main(){
  <<<flags definition>>>

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

<<<common procedures>>>

<<<first use case procedures>>>

<<<second use case procedures>>>
```

### First use case (Retrieve)

This use case is implemented by a procedure receiving the resource user names, and `PsFile` instance as parameters. It returns the last password for the supplied resource and user name. If the resource or user name doesn't exists it returns an error:

```go "first use case procedures"
func retrieve(resource, user string, pf PsFile) (r string, e error){
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
```

which import 

```go "imports" +=
"fmt"
```

### Second use case (Create)

The second use case is implemented by a procedure receiving the file, resource and user names, the `PsFile` instance, and the flag `easy` indicating when true that `apg` call is `apg -n 1`(generate an easy to remember password), or `apg -n 1 -a 1 -m 16` otherwise. It returns the created password and serializes the `PsFile` to a file with the supplied name, or returns an error. The name `easy` means. Also due to a shared operation between `create` and `retrieve`, getting the `[]Pass` associated in a `PsFile` to a a resource and user, the piece of code implementing it is refactored as `retrPs`.

```go "second use case procedures"

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
      Date: time.Now(),
      Password: r,
    })
    sr, e = yaml.Marshal(pf)
  }
  if e == nil {
    e = af.WriteFile(file, sr, os.ModePerm)
  }
  return
}

func retrPs(resource, user string, pf PsFile) (r []Pass, e error){
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
```

which import

```go "imports" +=
"github.com/spf13/afero"
"gopkg.in/yaml.v2"
"time"
"os/exec"
```

The `easy` flag is defined by:

```go "flags definition" +=
var easy bool
flag.BoolVar(&easy, "e", false, "Call 'apg -n 1' or 'apg -n 1 -a 1 -m 16'")
```

And having all parameters for calling `create` defined, and `create` implemented, which is the second and last use case of this program , the work is done.
