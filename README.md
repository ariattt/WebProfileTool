# WebProfileTool

A multi-threaded website profiling tool written in Go. Requests will be issued in batch.

## To Build
```
$ go build
```
A executeable named ```WebProfileTool``` should be generated under root

## To Run
``` 
$ ./WebProfileTool -url www.google.com/robots.txt -profile 100 -thread
********** Benchmark Result **********
------------------------------------
| Number of Requests   | 100       |
| Fastest Time         | 106 ms    |
| Slowest Time         | 305 ms    |
| Mean Time            | 133 ms    |
| Median Time          | 121 ms    |
| Precentage Succeeded | 100.00%   |
| Error Code Met       |           |
| Response Min Size    | 7645 byte |
| Response Max Size    | 7645 byte |
------------------------------------
```
Absent ```-profile``` flag will output server response to console.  
Help can be displayed with ```-h```
```
$ ./WebProfileTool -h
Usage of ./WebProfileTool:
  -profile int
    	Please specify the number of requests (default 1)
  -thread
    	Please indicate whether single or multi go-routine benchmark
    	* Single go-routine benchmark
    	  - first request slower than the followings
    	  - due to inevitable network system call cache
    	* Multi go-routine benchmark
    	  - requests issued in batch of GOMAXPROCS
    	  - first batch slower due to caching
  -url string
    	Please specify the endpoint to be profiled
```

## Design Decisions
The elapsed time measures the sum of establishing connection, sending http request, copying response to a local byte array and closing connection. 
A new connection is made for every request.

### Status Code
Status code 4xx and 5xx and number of occurence will be listed in Error Code Met section. Others including 1xx, 2xx, 3xx will be considered successful.

### Concurrency
The solution does now utilize go-routines. Each request is issued in a go routine. Go routines will be created in groups of GOMAXPROCS and it issues next group only if all go routines in the current group are finished.  
If all (e.g. 100) routines are issued at once, the go runtime will be drown and measured time will include a signicant portion of go routine wait time. Here's a specifix example to illustrate. *Routine A* sends request and the scheduler puts *Routine A* into sleep and runs other routines. The server could have responded to *Routine A* long before the scheduler yields control back to *Routine A*, making the measured time longer than the real time. 

### Request Headers
Here's the default http request header used. Feel free to change it if you need to. This is just from my Safari.
```
GET /robots.txt HTTP/1.1
HOST: www.google.com
Cache-Control: no-cache
Pragma: no-cache
```

### Robot.txt
The tool does not check against robot.txt, and does not sleep between calls, so please be careful using it! 
Watch out for Computer Abuse and Fraud Act (CAFA) or possibly being blacklisted!

## Findings
The tool is tested on 3 sites, ```www.google.com/robots.txt```, ```my-worker.maohuaw.workers.dev/``` and ```my-worker.maohuaw.workers.dev/links```, with 100 calls.
Here are the results.
```
$ ./WebProfileTool -url www.google.com/robots.txt -profile 100 -thread
********** Benchmark Result **********
------------------------------------
| Number of Requests   | 100       |
| Fastest Time         | 99 ms     |
| Slowest Time         | 313 ms    |
| Mean Time            | 127 ms    |
| Median Time          | 113 ms    |
| Precentage Succeeded | 100.00%   |
| Error Code Met       |           |
| Response Min Size    | 7645 byte |
| Response Max Size    | 7645 byte |
------------------------------------
$ ./WebProfileTool -url https://my-worker.maohuaw.workers.dev/ -profile 100 -thread
********** Benchmark Result **********
------------------------------------
| Number of Requests   | 100       |
| Fastest Time         | 148 ms    |
| Slowest Time         | 485 ms    |
| Mean Time            | 201 ms    |
| Median Time          | 173 ms    |
| Precentage Succeeded | 100.00%   |
| Error Code Met       |           |
| Response Min Size    | 2827 byte |
| Response Max Size    | 2827 byte |
------------------------------------
$ ./WebProfileTool -url https://my-worker.maohuaw.workers.dev/links -profile 100 -thread
********** Benchmark Result **********
-----------------------------------
| Number of Requests   | 100      |
| Fastest Time         | 148 ms   |
| Slowest Time         | 611 ms   |
| Mean Time            | 276 ms   |
| Median Time          | 161 ms   |
| Precentage Succeeded | 100.00%  |
| Error Code Met       |          |
| Response Min Size    | 889 byte |
| Response Max Size    | 889 byte |
-----------------------------------
```
To compare the speed of different calls, we are mostly interested in median, since network routing, DNS lookup and other network parts sometimes take a long time. The ```/``` call is slower than the ```/links``` call by 12 ms in median, because the former will call the latter on server end. Sometimes the first group of calls is slower than the rest, which I believe is caused by DNS caching. Http cache is manually disabled in the request header.  
The median is faster than mean implies the well-known long tail distribution of network calls. It also indicates consecutive calls are faster. This could be due to multiple factors including edge caching, server caching and router caching.

