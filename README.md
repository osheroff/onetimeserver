# onetimeserver

this is onetimeserver, a collection of utilities for spinning up
daemons(currently only mysql) in the context of a test process, and then
throwing them away at the end.

## how to use it

download the onetimeserver wrapper script:

```
curl https://raw.githubusercontent.com/osheroff/onetimeserver/master/onetimeserver > onetimeserver
```

In your test suite, execute the script:

```
require 'json'
server = JSON.parse(`onetimeserver -m 5.5`) # mysql 5.5 and 5.6 are currently supported
puts server
```

use the server for the life of your test suite.  It'll be killed and removed
after your suite exits.

## architecture

1.  shell script.  bootstraps the wrapper and the golang bits
2.  wrapper in C.  There because go can't seem to fork() properly.  forks and redirects
    STDOUT and STDERR to a file, then waits for a signal from the golang utility that the
    server has come up properly, and then exits, outputting information about the server in JSON.
3.  main utility, written in golang.  responsible for configuring and booting the server.


