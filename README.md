# WebProfileTool

A simple website profiling tool written in Go.

## To Build
```
$ go build
```
A executeable named ```WebProfileTool``` should be generated under root

## To Run
``` 
$ ./WebProfileTool -url www.google.com/robots.txt -profile 10
```
Absent ```-profile``` flag will output server response to console.
```-h``` will display help.
The result will be displayed similar to this.
```
********** Benchmark Result **********
------------------------------------
| Number of Requests   | 10        |
| Fastest Time         | 1029 ms   |
| Slowest Time         | 1079 ms   |
| Mean Time            | 1036 ms   |
| Median Time          | 1033 ms   |
| Precentage Succeeded | 100.00%   |
| Error Code Met       |           |
| Response Min Size    | 2204 byte |
| Response Max Size    | 2204 byte |
------------------------------------
```
## Design Decisions
The elapsed time measures the sum of establishing connection, sending http request, copying response to a local byte array and closing connection. 
A new connection is made for every request.

### Status Code
Status code 4xx and 5xx and number of occurence will be listed in Error Code Met section. Others including 1xx, 2xx, 3xx will be considered successful.

### Concurrency
The solution does not utilize go-routines, as time measued by concurrent threads will possibly include sleep or wait time. 
This is even more of a problem in case of network system calls. For instance, *Thread A* sends request and the scheduler puts *Thread A* into sleep and runs *Thread B*.
The server could have responded to *Thread A* long before the scheduler yields control back to *Thread A*, making the measured time longer than the real time.

### Request Headers
Here's the default http request header used. Feel free to change it if you need to. This is just from my Safari.
```
GET /robots.txt HTTP/1.1
HOST: www.google.com
Cache-Control: no-cache
Pragma: no-cache
Accept-Language: en-us
Connection: keep-alive
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
Accept-Encoding: gzip,deflate,br
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Safari/605.1.15

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
| Fastest Time         | 1027 ms   |
| Slowest Time         | 1103 ms   |
| Mean Time            | 1032 ms   |
| Median Time          | 1032 ms   |
| Precentage Succeeded | 100.00%   |
| Error Code Met       |           |
| Response Min Size    | 2204 byte |
| Response Max Size    | 2204 byte |
------------------------------------
$ ./WebProfileTool -url https://my-worker.maohuaw.workers.dev/ -profile 100
********** Benchmark Result **********
------------------------------------
| Number of Requests   | 100       |
| Fastest Time         | 1047 ms   |
| Slowest Time         | 1273 ms   |
| Mean Time            | 1062 ms   |
| Median Time          | 1051 ms   |
| Precentage Succeeded | 100.00%   |
| Error Code Met       |           |
| Response Min Size    | 2240 byte |
| Response Max Size    | 2436 byte |
------------------------------------
$ ./WebProfileTool -url https://my-worker.maohuaw.workers.dev/links -profile 100
********** Benchmark Result **********
-----------------------------------
| Number of Requests   | 100      |
| Fastest Time         | 1047 ms  |
| Slowest Time         | 1292 ms  |
| Mean Time            | 1058 ms  |
| Median Time          | 1051 ms  |
| Precentage Succeeded | 100.00%  |
| Error Code Met       |          |
| Response Min Size    | 775 byte |
| Response Max Size    | 785 byte |
-----------------------------------
```
The ```/``` call is slower than the ```/links``` call as expected, because the former will call the latter on server end. However, the margin is very small.
Sometimes the first call is slower than the rest, which I believe is caused by DNS caching. Http cache is manually disabled in the request header. 
The median is faster than mean indicates that consecutive calls are faster. This could be due to multiple factors including edge caching, server caching and router caching.
The measures time around 1000 milliseconds is clearly slower than the browser openning the corresponding webpage. 
I believe it is caused by copying response to a local byte array, which is limited by memory speed, and closing the connection. 
The browser also tends to preload the webpage before user clicks enter, which could save around 0.2 to 0.3 seconds.
