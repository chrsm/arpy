arpy
===

No, this has nothing to do with arp. It's "arrrrrrrr-py" as in `rpa` eg the format
that Ren'py uses to archive its files.


`arpy` is a really hacked together library/cmd for packing or unpacking RPAs (3, 3.2).


### Installation

If you have Go installed: `go install bits.chrsm.org/arpy/cmd/arpy@latest`

Otherwise, you can download [a release](https://github.com/chrsm/arpy/releases).


### Usage

```
$ arpy -h
rpa pack and unpack tool

Usage:
  arpy [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  pack        pack an RPA
  unpack      unpack an RPA

Flags:
  -h, --help         help for arpy
  -k, --key string   key for packing or unpacking - expect hex -> int (default "deadbeef")

Use "arpy [command] --help" for more information about a command.
```

pack:

```
$ arpy help pack
pack an RPA

Usage:
  arpy pack [flags]

Flags:
  -g, --glob string   glob pattern to include in archive (default "*")
  -h, --help          help for pack
  -o, --out string    RPA to create

Global Flags:
  -k, --key string   key for packing or unpacking - expect hex -> int (default "deadbeef")
```

unpack:

```
$ arpy help unpack
unpack an RPA

Usage:
  arpy unpack [flags]

Flags:
  -h, --help         help for unpack
  -i, --in string    RPA to extract
  -o, --out string   directory to write files to, defaults to /tmp (default "/tmp")

Global Flags:
  -k, --key string   key for packing or unpacking - expect hex -> int (default "deadbeef")
```


### Contributing

- don't be an ass
- write decent commit messages
- ???
- profit!

