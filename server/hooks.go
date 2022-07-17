package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/tidwall/gjson"
)

func transportNode(word string) string {
	url := "https://trial.api-service.navitime.biz/p2200470/v1/transport_node"
	request, _ := http.NewRequest(http.MethodGet, url, nil)
	params := request.URL.Query()
	params.Add("word", word)
	params.Add("signature", "47c0ac6e4851eb874422546c4db269476ec9cbef90ae95b8a9933e9375dff1d7")
	params.Add("request_code", "vzfmsK7brEY")
	request.URL.RawQuery = params.Encode()
	client := new(http.Client)
	res, _ := client.Do(request)
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	for i := 0; ; i++ {
		types := "items." + strconv.Itoa(i) + ".types"
		items := gjson.Get(string(body), types).String()
		if items == "" {
			break
		} else if strings.Contains(items, "station") {
			station := "items." + strconv.Itoa(i) + ".id"
			id := gjson.Get(string(body), station).String()
			return id
		}
	}

	return ""
}

func routeTransit(start string, goal string) string {
	url := "https://trial.api-service.navitime.biz/p2200470/v1/route_transit"
	request, _ := http.NewRequest(http.MethodGet, url, nil)
	params := request.URL.Query()
	params.Add("start", start)
	params.Add("goal", goal)
	params.Add("start_time", time.Now().Format("2006-01-02T15:04:05"))
	params.Add("signature", "47c0ac6e4851eb874422546c4db269476ec9cbef90ae95b8a9933e9375dff1d7")
	params.Add("request_code", "vzfmsK7brEY")
	request.URL.RawQuery = params.Encode()
	client := new(http.Client)
	res, _ := client.Do(request)
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	return string(body)

	// JSONを構造体にエンコード
	// station := "items.0.sections.0.name"
	// stationName := gjson.Get(string(body), station).String()

	// return (stationName)
	// file, _ := os.Create("response.json")
	// defer file.Close()
	// json.NewEncoder(file).Encode(response)
}

func timeToString(t time.Time) string {
	str := t.Format("2006-01-02 15:04:05")
	return str
}

func stringToTime(str string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", str)
	return t
}

func formatDateTime(src string) string {
	rep1 := strings.Replace(src, "+09:00", "", -1)
	rep2 := strings.Replace(rep1, "T", " ", -1)
	time := stringToTime(rep2).Add(time.Hour * +9)
	return strings.Replace(timeToString(time), "timeToString", "", -1)
}

func formatText(json string) string {
	buf := new(bytes.Buffer)
	buf.WriteString(fmt.Sprint("NAVITIME\n"))
	buf.WriteString(fmt.Sprint("```\n"))
	for i := 0; ; i++ {
		typeKey := "items.0.sections." + strconv.Itoa(i) + ".type"
		tyepVal := gjson.Get(json, typeKey).String()
		if strings.Contains(tyepVal, "point") {
			nameKey := "items.0.sections." + strconv.Itoa(i) + ".name"
			nameVal := gjson.Get(json, nameKey).String()
			buf.WriteString(fmt.Sprintf("✚%s\n", nameVal))
		} else if strings.Contains(tyepVal, "move") {
			fromTimeKey := "items.0.sections." + strconv.Itoa(i) + ".from_time"
			formTimeVal := gjson.Get(json, fromTimeKey).String()
			buf.WriteString(fmt.Sprintf("┃　　%s\n", formatDateTime(formTimeVal)))
			lineNameKey := "items.0.sections." + strconv.Itoa(i) + ".line_name"
			lineNameVal := gjson.Get(json, lineNameKey).String()
			buf.WriteString(fmt.Sprintf("┃　%s\n", lineNameVal))
			toTimeKey := "items.0.sections." + strconv.Itoa(i) + ".to_time"
			toTimeVal := gjson.Get(json, toTimeKey).String()
			buf.WriteString(fmt.Sprintf("┃　　%s\n", formatDateTime(toTimeVal)))
		} else {
			break
		}
	}
	buf.WriteString(fmt.Sprint("```"))
	return string(buf.Bytes())
}

func (p *Plugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
	reg1 := regexp.MustCompile(`ng( |　)+\S+( |　)+\S+`)

	if reg1.MatchString(post.Message) {
		reg2 := regexp.MustCompile(`( |　)+`)
		res := reg2.Split(post.Message, -1)
		p.API.LogDebug("Startdesu")
		start := transportNode(res[1])
		goal := transportNode(res[2])
		json := routeTransit(start, goal)
		route := formatText(json)
		post.Message = fmt.Sprint(route)
		p.API.LogDebug("Stop")
		return post, ""
	}

	// p.API.LogDebug("Qiita link is detected.")
	// post.Message = fmt.Sprintf("%s #Qiita", post.Message)
	// post.Hashtags = fmt.Sprintf("%s #Qiita", post.Hashtags)
	return post, ""
}
