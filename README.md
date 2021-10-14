# teeworlds-start

Small utility in order to start as many Teeworlds servers as you want.
This utility allows for a continuous availability of your servers even in case that they happen to crash due to some C++ bug.


## Requirements

[Go compiler installed](https://golang.org)

## Building the application

```
go build .

# or

make
```

## Usage

Execute all `autoexec_(teeworlds_srv)_XX.cfg` files that do have a corresponding `teeworlds_srv` located in the `executables` directory.
```
./teeworlds-start
```

In case you want to only start a group of servers with one instance of **teeworlds-start**, you can pass up to two command line arguments.
The first argument allows you to match executable file names and the second argument allows to further shrink the number of running servers by adding a regular expression to match the config files that you want to start with your matched executable.

Example:

The `executables` directory contains:
```
teeworlds_srv
zcatch_srv
gctf_srv
```

The `configs`directory contains:
```
autoexec_teeworlds_srv_peter-01.cfg
autoexec_teeworlds_srv_peter-02.cfg
autoexec_teeworlds_srv_peter-03.cfg

autoexec_zcatch_srv_peter-01.cfg
autoexec_zcatch_srv_01.cfg
autoexec_zcatch_srv_02.cfg

autoexec_gctf_srv_01.cfg
```

We do host a server for our friend Peter. And want his servers to be started by a different `teeworlds-start` application than ours.
We do not want our servers to go down when his servers need some updates of for whatever other reason you want to group your servers in a different way.

Peter has got three vanilla Teeworlds servers and a single zCatch server.

When we execute the following command, we start all of the configured servers:

```
./teeworlds-start
```

In order to start only zCatch servers we execute the next command:

```
./teeworlds-start zcatch_srv

or

./teeworlds-start zcatch
```
This starts Peter's and our `zcatch_srv` with every configuration that contains `autoexec_zcatch_srv_...`.


And finally we do want to startour own servers with a different starter than Peter's servers.
In order to do that, you need to provide a regular expression that matches all of the necessary executables followed by a regular expression that matches all of Peter's config files.

```
# to start Peter's servers
./teeworlds-start '.*' 'peter'

# to start our servers
./teeworlds-start '.*' '_\d+.cfg'
```

[Regular Expression for Peter](https://regex101.com/r/uCBtmP/1)  
[Regular Expression for us](https://regex101.com/r/H32Vwz/1)  


