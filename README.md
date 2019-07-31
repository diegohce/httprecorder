# httprecorder

<!-- [![Go Report Card](https://goreportcard.com/badge/github.com/diegohce/httprecorder)](https://goreportcard.com/report/github.com/diegohce/httprecorder) -->
[![GitHub release](https://img.shields.io/github/release/diegohce/httprecorder.svg)](https://github.com/diegohce/httprecorder/releases/)
[![Github all releases](https://img.shields.io/github/downloads/diegohce/httprecorder/total.svg)](https://github.com/diegohce/httprecorder/releases/)
[![GPLv3 license](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://github.com/diegohce/httprecorder/blob/master/LICENSE)
[![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)](https://github.com/diegohce/httprecorder/graphs/commit-activity)
[![HitCount](http://hits.dwyl.io/diegohce/httprecorder.svg)](http://hits.dwyl.io/diegohce/httprecorder)
[![Generic badge](https://img.shields.io/badge/deb%20package-no-red.svg)](https://github.com/diegohce/httprecorder/releases/)


## What is it?
httprecorder acts almost as a normal proxy. It records http requests and responses when running in recording mode (-record), and replays the stored responses when running in replay mode (-replay).

## But why?
I use it to run microservices (simple) integration tests. 

## Bind address

Default bind address and port: `0.0.0.0:8080`. It can be changed setting `HTTPRECORDER_BINDADDR` environment variable.


## Config file

`httprecorder.json` can be placed into project directory or, preferably, in `/etc/httprecorder`

* filename: Where to dump (or read from) the recorded content.
* default_host: Where every request that does not match one of `paths` will be routed to.
* paths: Where to route specific requests. If path ends with `/` it's interpreted as "begins with".

```json
{
	"filename": "recording.json",
	"default_host": {
		"host": "http://localhost:6666"
	},
	"paths": {
		"/badservice/status/400": {
			"host": "http://localhost:6666"
		},
		"/badservice/status/403": {
			"host": "http://localhost:6666"
		},
		"/badservice/status/404": {
			"host": "http://localhost:6666"
		},
		"/data/2.5/": {
			"host": "http://api.openweathermap.org:80"
		}
	}
}
```

# Status

httprecorder is still in a very early stage. There's code that can be improved for sure.


