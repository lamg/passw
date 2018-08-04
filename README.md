# Password manager (pmng)

The pmng password manager is a simple program for storing passwords associated to resources and user names in a text file using YAML format. It has two use cases:

## Retrieve
Retrieve a password given a file, resource and user name. A resource can be the Web site name which has the account. User and password have the usual meaning. In this case the command(having F as the file where information is stored, R as resource name, and U as user name) is:

```sh
pmng -r R -u U -f F
```

## Create
Create a password given a file, resource, user name, password and an `-e` flag which alternates between calling `apg` with the default parameters (this means it `apg -n 1` will generate a pronounceable password with length 8 when flag is absent) and calling `apg -n 1 -a 1 -m 16`, which will generate a random character password with length 16 when supplied. With R, U, F having the same above meaning and the `-c` flag distinguishing this call from the previous, the command is:

```
sh
pmng -c [-e] -r R -u U -f F
```