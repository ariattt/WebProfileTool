# WebProfileTool

A multi-threaded website profiling tool written in Go. Requests will be issued in batch.  
Highlights
* batch execution
* batch size dependent on underlying system
* TLS connection
* nice output format

## To Build
```
$ go build
```
A executeable named ```WebProfileTool``` should be generated under root

## To Run
``` 
$ ./WebProfileTool -url www.google.com/robots.txt -profile 100
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
  -dist
    	Print request time chronologically if flag set.
  -profile int
    	Please specify the number of requests (default 1)
  -thread int
    	Please specify the number of concurrent go-routine.
    	 -thread n, requests will be issued in groups of n.
    	    If n is invalid or absent, GOMAXPROCS will be used
    	
    	* Single go-routine benchmark
    	  - first request slower than the followings
    	  - inevitable network system call cache
    	* Multi go-routine benchmark
    	  - requests issued in batch of GOMAXPROCS
    	  - first batch slower due to caching (default -1)
  -url string
    	Please specify the endpoint to be profiled
```

## Design Decisions
The elapsed time measures the sum of establishing connection, sending http request, copying response to a local byte array and closing connection. 
A new connection is made for every request.

### Concurrency
The solution does now utilize go-routines. Each request is issued in a go routine. Go routines will be created in groups of GOMAXPROCS and it issues next group only if all go routines in the current group are finished.  
If all (e.g. 100) routines are issued at once, the go runtime will be drown and measured time will include a signicant portion of go routine wait time. Here's a specifix example to illustrate. *Routine A* sends request and the scheduler puts *Routine A* into sleep and runs other routines. The server could have responded to *Routine A* long before the scheduler yields control back to *Routine A*, making the measured time longer than the real time. 

### Status Code
Status code 4xx and 5xx and number of occurence will be listed in Error Code Met section. Others including 1xx, 2xx, 3xx will be considered successful.

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
$ ./WebProfileTool -url www.google.com/robots.txt -profile 100
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
$ ./WebProfileTool -url https://my-worker.maohuaw.workers.dev/ -profile 100 
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
$ ./WebProfileTool -url https://my-worker.maohuaw.workers.dev/links -profile 100
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
To compare the speed of different calls, we are mostly interested in median, since network routing, DNS lookup and other network parts sometimes take a long time. The ```/``` call is slower than the ```/links``` call by 12 ms in median, because the former will call the latter on server end.  
The median is faster than mean implies the well-known long tail distribution of network calls. It also indicates consecutive calls are faster. This could be due to multiple factors including edge caching, server caching and router caching.  
Network fluctuation is very observable. For example, the first group of calls is slower than the rest of groups. I believe it is caused by DNS caching. Http cache is manually disabled in the request header. Notice that the first 5 requests are of the same slow speed and one group in the middle is also slow at around 125 ms. All requests in side a group's speed fluctuate at the same pace is likely caused by low level delays.
```
$ ./WebProfileTool -url www.google.com/robots.txt -profile 100 -thread 5 -dist
Request times: 284 284 284 286 286 108 100 107 101 107 97 103 107 107 101 143 138 143 143 140 101 95 107 99 102 96 105 103 101 105 95 102 105 103 99 105 110 105 105 115 101 105 107 101 105 109 109 110 109 101 122 126 127 122 126 110 108 106 106 99 104 110 99 106 106 100 107 102 96 107 107 109 107 109 107 110 104 110 104 110 108 107 108 107 109 109 101 104 108 104 112 109 110 109 112 104 110 106 111 110 
(output omitted)
```


