package main

import (
    "bytes"
    "flag"
    "fmt"
    "io"
    "net"
    "sort"
    "strings"
    "time"
	"crypto/tls"
    "sync"
    "runtime"
)

type response struct{
    time int64
    code string
    size int
}

type item struct{
    name string
    value string
}

type options struct{
    n_req int
    url string
    n_thread int
    dist bool
}

var opts options

func min(a int64, b int64) int64 { if a<b {return a} else {return b}}
func max(a int64, b int64) int64 { if a>b {return a} else {return b}}

// return minTime, maxTime, meanTime, MedianTime....
func summarize(res []response) []string{
    if len(res) == 0 { return make([]string, 10)}

    if opts.dist{
        fmt.Printf("Request times: ")
        for _, resp := range res{
            fmt.Printf("%d ", resp.time)
        }
        fmt.Printf("\n")
    }
    
    sort.Slice(res, func(a,b int) bool{
        return res[a].time < res[b].time
    })
    var maxInt int64 = 9223372036854775807
    var sum,minSize,maxSize int64 = 0,maxInt, 0
    errCode := make(map[string]int)
    fail := 0
    for i:=0;i<len(res);i++{
        sum += res[i].time
        minSize = min(minSize, int64(res[i].size))
        maxSize = max(maxSize, int64(res[i].size))
        if strings.HasPrefix(res[i].code, "4") || strings.HasPrefix(res[i].code, "5"){
            fail += 1
            errCode[res[i].code] += 1
        }
    }
    fast,slow := res[0].time, res[len(res)-1].time
    median := res[len(res)/2].time
    if len(res) % 2 == 0{
        median = (res[len(res)/2].time + res[len(res)/2-1].time)/2
    }
    errString := ""
    for k := range errCode{
        errString += fmt.Sprintf("%s*%d, ", k, errCode[k])
    }
    if len(errString) > 2{
        errString = errString[:len(errString)-2]
    }
    summ := make([]string, 9)
    summ[0] = fmt.Sprintf("%d",len(res))
    summ[1] = fmt.Sprintf("%d ms", fast) 
    summ[2] = fmt.Sprintf("%d ms", slow) 
    summ[3] = fmt.Sprintf("%d ms", sum/int64(len(res))) 
    summ[4] = fmt.Sprintf("%d ms", median) 
    summ[5] = fmt.Sprintf("%.2f%%",(1.0 - float64(fail)/float64(len(res)))*100) 
    summ[6] = fmt.Sprintf("%s", errString) 
    summ[7] = fmt.Sprintf("%d byte", minSize) 
    summ[8] = fmt.Sprintf("%d byte", maxSize) 

    return summ
}

func report(res []response){
    summ := summarize(res) 
    items := make([]item, 9)
    items[0].name = "Number of Requests"
    items[1].name = "Fastest Time"
    items[2].name = "Slowest Time"
    items[3].name = "Mean Time"
    items[4].name = "Median Time"
    items[5].name = "Precentage Succeeded"
    items[6].name = "Error Code Met"
    items[7].name = "Response Min Size"
    items[8].name = "Response Max Size"

    for i:=0;i<9;i++{
        items[i].value = summ[i]
    }

    fmt.Printf("********** Benchmark Result **********\n")
    var leftLen, rightLen int64 = 0, 0
    for i:=0;i<9;i++{
        leftLen = max(leftLen, int64(len(items[i].name)))
        rightLen = max(rightLen, int64(len(items[i].value)))
    }
    width := int(leftLen + rightLen + 7);
    for i:=0;i<width;i++{
        fmt.Printf("-")
    }
    fmt.Printf("\n")
    for i:=0;i<9;i++{
        cur,_ := fmt.Printf("| %s", items[i].name)
        for ;cur<int(leftLen+3);cur++{
            fmt.Printf(" ")
        }
        cur,_ = fmt.Printf("| %s", items[i].value)
        for ;cur<int(rightLen+3);cur++{
            fmt.Printf(" ")
        }
        fmt.Printf("|\n")
    }
    for i:=0;i<width;i++{
        fmt.Printf("-")
    }
    fmt.Printf("\n")
 
}

func retrieve(host string, path string, buf *bytes.Buffer, success *bool){
    timeout, _ := time.ParseDuration("5s")
	d := net.Dialer{
		Timeout: timeout,
	}
    conn, err := tls.DialWithDialer(&d, "tcp", host + ":https", nil)
    if err != nil{
        fmt.Printf("conn error: %s\n", err)
        buf = nil
        return
    }
    defer conn.Close()

    fmt.Fprintf(conn, "GET "+path+" HTTP/1.0\r\n" + 
                        "HOST: "+host+"\r\n" +
                        "Cache-Control: no-cache\r\n" + 
                        "Pragma: no-cache\r\n" +
                        // "Accept-Language: en-us\r\n" +
                        // "Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8\r\n" +
                        // "Accept-Encoding: gzip,deflate,br\r\n" + 
                        "\r\n")

    _, err = io.Copy(buf, conn)
    if err != nil {
        fmt.Printf("read failed %v\n", err)
        buf = nil
        return
    }
    *success = true
}

func retrieve_wrapper(host, path string, res *response, wg *sync.WaitGroup){
    defer wg.Done()
    success := false
    var buf bytes.Buffer
    start := time.Now()
    retrieve(host, path, &buf, &success)
    elapsed := time.Now().Sub(start)

    code := "400"
    if success{
        str := buf.String()
        if ind := strings.Index(str, "\n"); ind != -1{
            str = str[:ind]
        }
        spts := strings.Split(str, " ")
        if len(spts) > 1{
            code = strings.Split(str, " ")[1]
        }else{
            code = "unknown"
            fmt.Printf(code)
        }
    }
    res.time = elapsed.Milliseconds()
    res.code = code
    res.size = buf.Len()
    if opts.n_req == 1{
        fmt.Printf("%s\n", buf.String())
    }
}

func benchmark(host string, path string){

    res := make([]response, opts.n_req)
    var wg sync.WaitGroup
    maxprocs := opts.n_thread
    if maxprocs <= 0 {
        maxprocs = runtime.GOMAXPROCS(-1)
    }
    // fmt.Printf("%d\n", maxprocs)
    for i:=0;i<opts.n_req;i++{
        wg.Add(1)
        go retrieve_wrapper(host, path, &res[i], &wg)
        if i % maxprocs == maxprocs-1 {
            wg.Wait()
        }
    }
    wg.Wait()
    report(res)
}

// prase url and return host and path
func parse_url() (string, string){
    url := opts.url
    if ind := strings.Index(url, "://"); ind != -1{
        url = url[ind+3:]
    }

    host, path := url, "/"
    if ind := strings.Index(url, "/"); ind != -1{
        host = url[:ind]
        path = url[ind:]
    }

    return host, path
}

func main() {
    flag.StringVar(&opts.url, "url", "", "Please specify the endpoint to be profiled")
    flag.IntVar(&opts.n_req, "profile", 1, "Please specify the number of requests")
    flag.IntVar(&opts.n_thread, "thread", -1, "Please specify the number of concurrent go-routine.\n" + 
                                                " -thread n, requests will be issued in groups of n.\n" +
                                                "    If n is invalid or absent, GOMAXPROCS will be used\n" + 
                                                "\n" +
                                                "* Single go-routine benchmark\n" +
                                                "  - first request slower than the followings\n" + 
                                                "  - inevitable network system call cache\n" + 
                                                "* Multi go-routine benchmark\n" +
                                                "  - requests issued in batch of GOMAXPROCS\n" + 
                                                "  - first batch slower due to caching",
                                            )
    flag.BoolVar(&opts.dist, "dist", false, "Print request time chronologically if flag set.")
    flag.Parse()

    host, path := parse_url()

    benchmark(host, path)
}