/*
 * Simple URL's security score check tool.
 *
 * @author    yasutakatou
 * @copyright 2020 yasutakatou
 * @license   MIT License
 */
package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"gopkg.in/ini.v1"
	"net/http"
	"strings"
	"strconv"
	"github.com/rocketlaunchr/google-search"
	"context"
	"net"
	"time"
	"github.com/nsf/termbox-go"
	"errors"
	"log"
	"encoding/json"
	"encoding/base64"
	"math/rand"
)

type responseData struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type scoreData struct {
	CACHE  int64
	pCACHE string
	DNS    []string
	SEARCH int
	SSL    []string
}

type historyData struct {
	URL string
	TIME int64
	score string
}

var (
	rs1Letters     = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	Debug bool
	score scoreData
	history []historyData
	Token string
)

func init() {
	rand.Seed(time.Now().UnixNano())
	Token = RandStr(8)
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.Flush()

	_Debug := flag.Bool("debug", false, "[-debug=debug mode (true is enable)]")
	_cert := flag.String("cert", "localhost.pem", "[-cert=ssl_certificate file path (if you don't use https, haven't to use this option)]")
	_key := flag.String("key", "localhost-key.pem", "[-key=ssl_certificate_key file path (if you don't use https, haven't to use this option)]")
	_port := flag.String("port", "8080", "[-port=port number]")
	_config := flag.String("config", "config", "[-config=config file name]")

	flag.Parse()

	Debug = bool(*_Debug)	

	loadConfig(string(*_config))

	go func() {
		http.HandleFunc("/token", startHandler)

		http.HandleFunc("/"+Token+"/api/",func(w http.ResponseWriter, r *http.Request) {
			serverName,err := base64.URLEncoding.DecodeString(r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:])
			if err == nil {
				if Debug == true {
					fmt.Println("api call: " + r.RemoteAddr + " " + string(serverName))
				}
				apiHandler(w, r, string(serverName))
			}
		})

		err := http.ListenAndServeTLS(":"+string(*_port), string(*_cert), string(*_key), nil)
		if err != nil {
			log.Fatal("ListenAndServeTLS: ", err)
		}
	}()
	fmt.Println("[ Press enter or space key. Server Status display. Escape is exit.]")
	startServer(string(*_port))

	saveConfig(string(*_config))
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if Debug == true {
		fmt.Println("token call: ", r.RemoteAddr, r.URL.Path)
	}

	data := &responseData{Status: "Success", Message: Token}
	outputJson, err := json.Marshal(data)
	if err != nil {
		fmt.Println("%s")
		return
	}

	w.Write(outputJson)
}

func saveConfig(filename string) bool {
	if len(filename) == 0 {
		return false
	}

	file, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer file.Close()

	writeFile(file, "[CACHE]")
	writeFile(file, score.pCACHE)

	writeFile(file, "[DNS]")
	for i := 0; i < len(score.DNS); i++ {
		writeFile(file, score.DNS[i])
	}

	writeFile(file, "[SEARCH]")
	writeFile(file, strconv.Itoa(score.SEARCH))

	writeFile(file, "[SSL]")
	for i := 0; i < len(score.SSL); i++ {
		writeFile(file, score.SSL[i])
	}

	writeFile(file, "[HISTORY]")
	for i := 0; i < len(history); i++ {
		writeFile(file, history[i].URL+" "+strconv.FormatInt(history[i].TIME, 10)+" "+history[i].score)
	}

	return true
}

func writeFile(file *os.File, strs string) bool {
	_, err := file.WriteString(strs + "\n")
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func checkHistory(serverName string) (string) {
	t := time.Now()
	for i := 0; i < len(history); i++ {
		if serverName == history[i].URL {
			if t.Unix() < history[i].TIME {
				return history[i].score
			}
		}
	}
	return ""
}

func printStar(strs string) (string) {
	buff := ""
	for _, c := range strs { 
		if string([]rune{c}) == "0" {
			buff = buff + "★"
		} else {
			buff = buff + "☆"
		}
	}
	return buff
}

func apiHandler(w http.ResponseWriter, r *http.Request, serverName string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Content-Type", "application/json")

	scores := checkHistory(serverName)

	if len(scores) == 0 {
		if matchDNS(serverName) == true {
			scores = scores + "0"
		} else {
			scores = scores + "1"
		}
	
		if matchSearch(serverName) == true {
			scores = scores + "0"
		} else {
			scores = scores + "1"
		}
	
		if matchSSL(serverName) == true {
			scores = scores + "0"
		} else {
			scores = scores + "1"
		}

		t := time.Now()
		history = append(history, historyData{URL: serverName, TIME: (t.Unix() + score.CACHE), score: scores})
	}

	fmt.Println("Securescore: " + printStar(scores))

	data := &responseData{Status: "Success", Message: scores}
	outputJson, err := json.Marshal(data)
	if err != nil {
		fmt.Println("%s")
		return
	}
	w.Write(outputJson)
}


func startServer(port string) {
	termbox.SetInputMode(termbox.InputEsc)

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case 13, 32: //Enter, Space
				_, ip, err := getIFandIP()
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("source ip: ", ip, " port: ", port)
				}
			case 27: //Escape
				termbox.Flush()
				return
			default:
			}
		}
	}
}

// FYI: https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
func getIFandIP() (string, string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return iface.Name, ip.String(), nil
		}
	}
	return "", "", errors.New("are you connected to the network?")
}

func matchDNS(url string) bool {
	for i := 0; i < len(score.DNS); i++ {
		if DNSLookup(score.DNS[i], cleanURL(url)) == true {
			return true
		}
	}
	return false
}

func DNSLookup(DNSServer, url string) bool {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, "udp", DNSServer + ":53")
		},
	}
	ip, _ := r.LookupHost(context.Background(), url)
	if ip != nil {
		if len(ip[0]) > 6 {
			if Debug == true {
				fmt.Println(" -- [DNS]: " + DNSServer + " --")
				fmt.Println(ip[0])
			}
			return true
		}
	}
	return false
}


func matchSearch(url string) bool {
	ctx := context.Background()
	result,err := googlesearch.Search(ctx, cleanURL(url))

	if err != nil {
		fmt.Println(err)
		return false
	}

	fmt.Println(" -- [SEARCH]: " + url + " --")
	if Debug == true {
		for i := 0; i < len(result); i++ {
			fmt.Println(strconv.Itoa(result[i].Rank) + " " + result[i].URL)
		}
	}

	for i := 0; i < len(result); i++ {
		if strings.Index(result[i].URL, url) != -1 && result[i].Rank <= score.SEARCH {
			return true
		}
	}
	return false
}

func cleanURL(tmpStr string) (string) {
	tmpStr = strings.Replace(tmpStr, "https://", "", -1)
	tmpStr = strings.Replace(tmpStr, "http://", "", -1)
	return strings.Split(tmpStr, "/")[0]
}

func matchSSL(url string) bool {
	if strings.Index(url, "http://") != -1 {
		return false
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer resp.Body.Close()
	
	strs := fmt.Sprintf("%s",resp.TLS.PeerCertificates[0].Issuer)
	if Debug == true {
		fmt.Println(" -- [SSL]: " + url + " --")
		fmt.Println(strs)
	}

	for i := 0; i < len(score.SSL); i++ {
		if strings.Index(strs, score.SSL[i]) != -1 {
			return true
		}
	}
	return false
}

func loadConfig(filename string) {
	loadOptions := ini.LoadOptions{}
	loadOptions.UnparseableSections = []string{"CACHE","DNS", "SEARCH", "SSL", "HISTORY"}

	cfg, err := ini.LoadSources(loadOptions, filename)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	setCache(&score.CACHE, "CACHE", cfg.Section("CACHE").Body())
	setSingleConfigStr(&score.DNS, "DNS", cfg.Section("DNS").Body())
	setSingleConfigInt(&score.SEARCH, "SEARCH", cfg.Section("SEARCH").Body())
	setSingleConfigStr(&score.SSL, "SSL", cfg.Section("SSL").Body())
	setHistorys("HISTORY", cfg.Section("HISTORY").Body())
}

func setCache(config *int64, configType, datas string) {
	if Debug == true {
		fmt.Println(" -- " + configType + " --")
	}
	for _, v := range regexp.MustCompile("\r\n|\n\r|\n|\r").Split(datas, -1) {
		if len(v) > 0 {
			tabs := strings.Split(v, " ")
			count, err := strconv.Atoi(tabs[0])
			if err != nil {
				return
			}	
			unit := setUnit(tabs[1])
			if unit != 0 {
				*config = int64(count) * unit
			}
			score.pCACHE = v
		}
		if Debug == true {
			fmt.Println(v)
		}
	}
}

func setUnit(unit string) (int64) {
	switch unit {
	case "d", "D":
		return 12 * 60 * 60
	case "h", "H":
		return 60 * 60
	case "m", "M":
		return 60
	case "s", "S":
		return 1
	default:
		return 0
	}
	return 0
}

func setSingleConfigInt(config *int, configType, datas string) {
	if Debug == true {
		fmt.Println(" -- " + configType + " --")
	}
	for _, v := range regexp.MustCompile("\r\n|\n\r|\n|\r").Split(datas, -1) {
		if len(v) > 0 {
			tmp, err := strconv.Atoi(v)
			if err == nil {
				*config = tmp
			}		
		}
		if Debug == true {
			fmt.Println(v)
		}
	}
}

func setSingleConfigStr(config *[]string, configType, datas string) {
	if Debug == true {
		fmt.Println(" -- " + configType + " --")
	}
	for _, v := range regexp.MustCompile("\r\n|\n\r|\n|\r").Split(datas, -1) {
		if len(v) > 0 {
			*config = append(*config, v)
		}
		if Debug == true {
			fmt.Println(v)
		}
	}
}

func setHistorys(configType, datas string) {
	t := time.Now()
	if Debug == true {
		fmt.Println(" -- " + configType + " --")
	}
	for _, v := range regexp.MustCompile("\r\n|\n\r|\n|\r").Split(datas, -1) {
		if len(v) > 0 {
			tabs := strings.Split(v, " ")
			
			times, _ := strconv.ParseInt(tabs[1], 10, 64)
			if t.Unix() < times {
				history = append(history, historyData{URL: tabs[0], TIME: times, score: tabs[2]})
			}
		}
	}
	if Debug == true {
		fmt.Println(history)
	}
}

func RandStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = rs1Letters[rand.Intn(len(rs1Letters))]
	}
	return string(b)
}
