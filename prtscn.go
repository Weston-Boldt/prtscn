// https://medium.com/@KentGruber/building-a-high-performance-port-scanner-with-golang-9976181ec39d

package main

import (
    "context"
    "fmt"
    "net"
    "os/exec"
    "strconv"
    "strings"
    "sync"
    "time"
    "flag"

    "golang.org/x/sync/semaphore"
)

type PortScanner struct {
    ip string
    lock *semaphore.Weighted
}

func Ulimit() int64 {
    out, err := exec.Command("ulimit", "-n").Output()

    if err != nil {
        panic(err)
    }

    s := strings.TrimSpace(string(out))
    i, err := strconv.ParseInt(s, 10, 64)

    if err != nil {
        panic(err)
    }

    return i
}

func ScanPort(ip string, port int, timeout time.Duration) {
    target := fmt.Sprintf("%s:%d", ip, port)
    conn, err := net.DialTimeout("tcp", target, timeout)

    if err != nil {
        if strings.Contains(err.Error(), "too many open files") {
            time.Sleep(timeout)
            ScanPort(ip, port, timeout)
        } else {
            fmt.Println(port, "closed")
        }
        return
    }

    conn.Close()
    fmt.Println(port, "open")
}

func (ps *PortScanner) Start(f, l int, timeout time.Duration) {
    wg := sync.WaitGroup{}
    defer wg.Wait()

    for port := f; port <= l; port++ {
        wg.Add(1)
        ps.lock.Acquire(context.TODO(), 1)

        go func(port int) {
            defer ps.lock.Release(1)
            defer wg.Done()
            ScanPort(ps.ip, port, timeout)
        }(port)
    }
}

func main() {

    ipPtr := flag.String("ip", "127.0.0.1", "IP address")
    startPortPtr := flag.Int("start", 1, "starting port")
    endPortPtr := flag.Int("end", 65535, "end port")
    timeoutPtr := flag.Int("timeout", 500, "timeout in miliseconds")

    flag.Parse()

    ps := &PortScanner{
        ip: *ipPtr,
        lock: semaphore.NewWeighted(Ulimit()),
    }

    ps.Start(*startPortPtr, *endPortPtr, time.Duration(*timeoutPtr)*time.Millisecond)
}
