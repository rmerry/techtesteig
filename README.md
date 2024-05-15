# Eiger Coding Challenge - Richard Merry

## Overview

I chose to use the reference [bitcoin core](https://github.com/bitcoin/bitcoin) implementation for this challenge which I have added to this repository as a [git submodule](https://git-scm.com/book/en/v2/Git-Tools-Submodules).

I recognise there are some issues still but I've spent as much time as I can on it. Hopefully if gives a clear indication of the direction I was looking to take it. Ideally I would add tests to the client using a server but I ran out of time. 

## Running locally

I have included the [bitcoin core](https://github.com/bitcoin/bitcoin) as a [git submodule](https://git-scm.com/book/en/v2/Git-Tools-Submodules) in this repository---this code lives in the `bitcoin` directory. Instructions for building this for your particular environment can be found in the `bincoin/doc` folder.

For testing the Go client we can run the bitcoin core daemon in regression mode[1]:

```
> bitcoind -regtest -daemon
```

or simply

```
make run-bitcoind
```
The latter assumes that bitcoind has been compiled and lives in the default build location.

We can then run the Go client:

```
make run
```

To stop the client you can issue an interrupt with `ctrl+c`.

## Testing

A small number of unit tests are provided and can be run with:

```
make test
```

### References

[1] https://developer.bitcoin.org/examples/testing.html#regtest-mode
