This data collection package is an early version that I wrote for my own personal needs. There is a significantly better (way more tests, no global state, and focused on being implemented in the cloud) version that I use in production, but I'm keeping that under wraps for now. If you so desire, feel free to use as you see fit, I'm posting this to show people that I do infact know how to code. That being said, it certainly works well enough to maintain connections to decentralized and centralized exchanges.

This package has the tools to reboot predefined processes, manage websocket connections, standardize data, and dump data into a database (I'm using influxdb, but others can be plugged in with a little work).

requires influxDB's github.com/influxdata/influxdb1-client/v2 be installed