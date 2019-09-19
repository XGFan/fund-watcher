package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jedib0t/go-pretty/table"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type FundId struct {
	Id     string
	Weight float64
}
type FundWrapper struct {
	Fund Fund `json:"Expansion"`
}
type Fund struct {
	FCODE     string
	SHORTNAME string
	GZTIME    string
	GSZZL     string
}

type FundItem struct {
	Fund
	Weight float64
}

type FundResult struct {
	Funds []FundItem
	Avg   float64
}

func NewFundResult(items []FundItem) FundResult {
	fundResult := FundResult{}
	fundResult.Funds = items
	fundResult.Avg = fundResult.avg()
	return fundResult
}

func (p FundResult) avg() float64 {
	total := 0.0
	totalWeight := 0.0
	for _, fund := range p.Funds {
		total = total + fund.Zzl()*fund.Weight
		totalWeight = totalWeight + fund.Weight
	}
	return total / totalWeight
}

func (p FundItem) Zzl() float64 {
	zzl, _ := strconv.ParseFloat(p.GSZZL, 64)
	return zzl
}

func main() {
	isServer := flag.Bool("s", false, "server mode")
	port := flag.Int("p", 16000, "port number, only valid in server mode")
	flag.Parse()
	if *isServer {
		Webserver(*port)
	} else {
		Console()
	}
}

func Console() {
	ids := LoadFundIds()
	fundResult := GetFundResult(ids)
	t := table.NewWriter()
	for _, fund := range fundResult.Funds {
		t.AppendRow([]interface{}{fund.FCODE, fund.SHORTNAME, fund.GZTIME, fund.GSZZL, fund.Weight})
	}
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "名称", "时间", "估算增长率", "权重"})
	avg := fmt.Sprintf("%f%s", fundResult.avg(), "%")
	t.AppendFooter(table.Row{"Avg", "", "", avg, ""})
	t.Render()
}

func Webserver(port int) {
	http.HandleFunc("/fund", handler)
	addr := fmt.Sprintf(":%d", port)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	ids := LoadFundIds()
	result := GetFundResult(ids)
	bytes, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprintln(w, string(bytes))
}

func LoadFundIds() []FundId {
	home := os.Getenv("HOME")
	bytes, _ := ioutil.ReadFile(home + "/.fund")
	s := string(bytes)
	split := strings.Split(s, "\n")
	var fundIds []FundId
	for _, v := range split {
		i := strings.Split(v, ",")
		weight, _ := strconv.ParseFloat(i[1], 64)
		x := FundId{i[0], weight}
		fundIds = append(fundIds, x)
	}
	return fundIds
}
func GetFundResult(fundIds []FundId) FundResult {
	var fundItems []FundItem
	for _, idAndWeight := range fundIds {
		fund := GetFund(idAndWeight.Id)
		item := FundItem{fund, idAndWeight.Weight}
		fundItems = append(fundItems, item)
	}
	return NewFundResult(fundItems)
}

func GetFund(id string) Fund {
	req, _ := http.NewRequest("GET",
		"https://fundmobapi.eastmoney.com/FundMApi/FundVarietieValuationDetail.ashx",
		nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Mobile/15A372 Safari/604.1")
	req.Header.Set("Accept", "application/json")
	q := req.URL.Query()
	q.Add("FCODE", id)
	q.Add("RANGE", "y")
	q.Add("deviceid", "Wap")
	q.Add("plat", "Wap")
	q.Add("product", "EFund")
	q.Add("version", "2.0.0")
	ts := fmt.Sprintf("%d", time.Now().UnixNano()/1000)
	q.Add("_", ts)
	req.URL.RawQuery = q.Encode()
	resp, _ := http.DefaultClient.Do(req)
	bytes, _ := ioutil.ReadAll(resp.Body)
	wrapper := &FundWrapper{}
	e := json.Unmarshal(bytes, wrapper)
	if e != nil {
		log.Fatal(e)
	}
	return wrapper.Fund
}
