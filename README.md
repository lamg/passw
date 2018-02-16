# Passw

Passw is a simple password manager which stores passwords in a YAML file encrypted with GPG, using the default key. The use cases are the following, having specified a resource (-r option) (which is something that holds several user accounts, like Facebook, GitHub or GPG if an account means an individual key), an user name (-u option) and a file (-f option):

## Create resource
Creates a resource with the specified resource name, in case it exists the user and password will be appended to the resource's accounts. The user may choose to generate (-g flag)a new password using `apg`, which by default is called with `-m 16 -a 1 -n 1` arguments. If -e flag is passed then the arguments to `agp` are `-n `. If -g flag isn't present, then the password is the string passed with -p option.

## Delete resource

## Get resource