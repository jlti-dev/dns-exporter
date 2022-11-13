package main

import (
        "fmt"
        "time"
        "net/http"
        "os"
        "os/signal"
        "syscall"
        "github.com/miekg/dns"
        "github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promhttp"
        "github.com/prometheus/client_golang/prometheus/promauto"
)
type server struct{
        name string
        group string
        url string
        class string
        dns string
        expectedIp string
        rtt int64
        exp int
        err int
}
var metric_rtt = promauto.NewGaugeVec( prometheus.GaugeOpts{
        Namespace: "dns",
        Name: "request_duration",
        Help: "Length of request duration in Microseconds",
}, []string{"host","name","group","class","dns","expectedIp"})
var metric_exp =promauto.NewGaugeVec( prometheus.GaugeOpts{
        Namespace: "dns",
        Name: "ip_matched",
        Help: "compares all found ips to the expected IP and is 1 if the expected IP is found",
}, []string{"host","name","group","class","dns","expectedIp"})
var metric_err = promauto.NewGaugeVec( prometheus.GaugeOpts{
        Namespace: "dns",
        Name: "error_in_resolution",
        Help: "1 if there is an error, 0 else",
}, []string{"host","name","group","class","dns","expectedIp"})
var metrics []server
var dns_client = dns.Client{}
var stop = false
func main(){
        readFile("checkHosts")

        if len(metrics) == 0 {
                fmt.Println("No Servers found!")
                return
        }
        for _, v := range metrics{
                fmt.Println(v)
                go runCheck(v, 10)
        }

        fmt.Println("Starting Exporter")
        http.Handle("/metrics", promhttp.Handler())
        go http.ListenAndServe(":8080", nil)

        os_call := make(chan os.Signal, 1)
        signal.Notify(os_call, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM,syscall.SIGQUIT)

        for {
                select {
                        case <-os_call:
                                fmt.Println("Got interrupt")
                                stop = true
                                return
                }
        }
}
func runCheck(s server, sleep time.Duration){
        for {
                if stop == true{
                        fmt.Printf("Terminating URL %s with DNS %s\n", s.url, s.dns)
                        return
                }
                s.rtt, s.err, s.exp = checkServer(s.url, s.dns, s.expectedIp)
                metric_rtt.WithLabelValues(s.url, s.name, s.group, s.class, s.dns, s.expectedIp).Set(float64(s.rtt))
                metric_exp.WithLabelValues(s.url, s.name, s.group, s.class, s.dns, s.expectedIp).Set(float64(s.exp))
                metric_err.WithLabelValues(s.url, s.name, s.group, s.class, s.dns, s.expectedIp).Set(float64(s.err))
                time.Sleep(sleep * time.Second)
        }

}
func checkServer(url string, server string, expectedIp string)(int64, int, int){
        m := dns.Msg{}
        m.SetQuestion(url+".", dns.TypeA)
        r,t,err := dns_client.Exchange(&m, server+":53")
        found := ""
        if err != nil {
                fmt.Printf("URL %s für DNS %s hat einen schwerwiegenden Fehler!\n", url, server)
                fmt.Println(err)
                return 1000, 1, 0
        }
        if len(r.Answer) == 0{
                fmt.Printf("DNS %s kann URL %s nicht auflösen", server, url)
                return 1000, 1, 0
        }
        //Nur matchen, wenn auch eine expected IP angegeben ist.
        if len(expectedIp) > 0 {
                for _, ans := range r.Answer {
                        aRecord := ans.(*dns.A)
                        if aRecord.A.String() == expectedIp && len(aRecord.A.String()) > 0 {
                                return int64(t / time.Microsecond), 0, 1
                        }
                        found = aRecord.A.String()
                }
        }
        fmt.Printf("Expected IP was: %s, but found %s\n", expectedIp, found)
        return int64(t / time.Microsecond), 0, 0
}
