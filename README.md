# gweather

gweather is CLI tool for acquiring weather information regularly.

Weather information is obtained from [Meteorological Agency](http://xml.kishou.go.jp/xmlpull.html) data and stored on redis.


## Install

```shell
go get github.com/hlts2/gweather
```

## Example

Get data at one second intervals and store in redis.
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

## Contents stored in redis

```
$ redis-cli -p 1111 --raw
127.0.0.1:1111> keys *
気象警報・注意報_盛岡地方気象台
気象特別警報・警報・注意報_盛岡地方気象台
127.0.0.1:1111> get 気象特別警報・警報・注意報_盛岡地方気象台
```

Contents when acquired with the key(`気象特別警報・警報・注意報_盛岡地方気象台`)
https://github.com/hlts2/gweather/blob/master/_data/data.json







