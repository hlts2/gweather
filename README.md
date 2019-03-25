# gweather

gweather is CLI tool for acquiring weather information regularly.

Weather information is obtained from [Meteorological Agency](http://xml.kishou.go.jp/xmlpull.html) data and stored on redis


## Install

```shell
go get github.com/hlts2/gweather
```

## Example

```
$ gweather -s 1 --host redis://127.0.0.1:6379

```

## Usage

```
$ gweather --help
CLI tool for acquiring weather information regularly

Usage:
  gweater [flags]

Flags:
  -h, --help          help for gweater
      --host string   Host address for Redis (default "redis://127.0.0.1:6379")
  -s, --second uint   Interval to get weather information (default 180)
      --version       version for gweater
```
