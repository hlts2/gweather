# e2e

## Files

- [main.go](main.go) - example source code

## Requirements

To run this example you will need Docker and docker installed. See installation guide at https://docs.docker.com/install/

## Example

## Start redis container

```bash
$ docker pull redis:5.0.4-alpine
$ docker run -p 1111:6793 redis:5.0.4-alpine
```

## Start CLI application
```
$ go run main.go -s 1 --host redis://127.0.0.1:1111
2019-03-25 23:04:59	[INFO]:	Start cli application
2019-03-25 23:05:00	[INFO]:	Start job to get information
2019-03-25 23:05:00	[INFO]:	Finish job. time: 313.033389ms
2019-03-25 23:05:01	[INFO]:	Start job to get information
2019-03-25 23:05:01	[INFO]:	Finish job. time: 69.610392ms
....
```
