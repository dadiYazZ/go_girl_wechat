package main

import (
	"encoding/json"
	"fmt"
	"github.com/robfig/cron"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	APPID           string
	APPSECRET       string
	templateid      string
	openid          string
	city            string
	love_start_time string
	spec            string
)

type token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type weather struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Address       string `json:"address"`
		CityCode      string `json:"cityCode"`
		Temp          string `json:"temp"`
		Weather       string `json:"weather"`
		WindDirection string `json:"windDirection"`
		WindPower     string `json:"windPower"`
		Humidity      string `json:"humidity"`
		ReportTime    string `json:"reportTime"`
	} `json:"data"`
}

func main() {

	loadConfig()

	c := cron.New()
	//spec := "*/5 * * * * ?"
	//spec := "0 0 7 * * *"

	c.AddFunc(spec, func() { // AddFunc 是添加任务的地方，此函数接收两个参数，第一个为表示定时任务的字符串，第二个为真正的真正的任务。
		var weatherData weather = getWeather()

		timeData := getHourDiffer(love_start_time, time.Now().Format("2006-01-02 15:04:05"))

		Templatepost(weatherData.Data.Weather, weatherData.Data.Temp, getEarthy(), timeData/24, timeData)
	})
	c.Start()

	fmt.Println("已经开始运行啦~")

	for {
		loadConfig()
	}

}

func loadConfig() {
	cfg, err := ini.Load("config.ini")

	if err != nil {
		fmt.Println("读取配置文件出错!")
		return
	}
	APPID = cfg.Section("wechat").Key("appid").String()
	APPSECRET = cfg.Section("wechat").Key("app_secret").String()
	templateid = cfg.Section("wechat").Key("template_id").String()
	templateid = cfg.Section("wechat").Key("template_id").String()
	openid = cfg.Section("wechat").Key("openid").String()
	love_start_time = cfg.Section("setting").Key("love_time").String()
	city = cfg.Section("setting").Key("city").String()
	spec = cfg.Section("cron").Key("spec").String()

}

// 获取相差时间
func getHourDiffer(start_time, end_time string) int64 {
	var hour int64
	t1, err := time.ParseInLocation("2006-01-02 15:04:05", start_time, time.Local)
	t2, err := time.ParseInLocation("2006-01-02 15:04:05", end_time, time.Local)
	if err == nil && t1.Before(t2) {
		diff := t2.Unix() - t1.Unix() //
		hour = diff / 3600
		return hour
	} else {
		return hour
	}
}

// 获取土味情话
func getEarthy() string {
	resp, err := http.Get("https://api.lovelive.tools/api/SweetNothings")

	defer resp.Body.Close()

	if err != nil {
		fmt.Println("获取每日一句失败", err)
		return fmt.Sprintf("%s", err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("读取内容失败", err)
		return "读取内容失败"
	}

	return string(body)

}

// 发送模板消息
func Templatepost(weatherData, temperature, data string, dayTime, dayHours int64) {
	rand.Seed(time.Now().UnixNano())

	accessTokenData := getaccesstoken()

	dayTimeData := strconv.FormatInt(dayTime, 10)
	dayHoursData := strconv.FormatInt(dayHours, 10)

	colorArray := []string{"#6e2dc0", "#8ead8b", "#ad191f", "#261e8b", "#FF4500", "#7FFF00", "#F08080", "#FF6347", "#FFA07A", "#A9A9A9"}

	number := rand.Intn(len(colorArray) - 1)

	reqdata := "{\"weather\":{\"value\":\"" + weatherData + "\"}, \"temperature\":{\"value\":\"" + temperature + "\"}  ,\"data\":{\"value\":\"" + data + "\" , \"color\":\"" + colorArray[number] + "\"} , \"dayTime\":{\"value\":\"" + dayTimeData + "\"} , \"dayHours\":{\"value\":\"" + dayHoursData + "\"} }"

	templatepost(accessTokenData, reqdata, "", templateid, openid)

}

// 获取天气
func getWeather() weather {
	url := fmt.Sprintf("https://www.mxnzp.com/api/weather/current/%v?app_id=%v&app_secret=%v", city, "iloophp0hm0il7nq", "Zk50bnJmU0tEQ0VSVVlrUnVUTU81UT09")
	resp, err := http.Get(url)

	if err != nil {
		fmt.Println("获取天气失败~", err)
		return weather{}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("解析天气失败~", err)
		return weather{}
	}

	weather := weather{}
	json.Unmarshal(body, &weather)

	return weather

}

// 获取微信accesstoken
func getaccesstoken() string {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%v&secret=%v", APPID, APPSECRET)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("获取微信token失败", err)
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("微信token读取失败", err)
		return ""
	}

	token := token{}
	err = json.Unmarshal(body, &token)
	if err != nil {
		fmt.Println("微信token解析json失败", err)
		return ""
	}

	return token.AccessToken
}

// 发送模板消息
func templatepost(access_token string, reqdata string, fxurl string, templateid string, openid string) {
	url := "https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=" + access_token

	//reqbody := "{\"touser\":\"" + openid + "\", \"template_id\":\"" + templateid + "\", \"url\":\"" + fxurl + "\", \"data\": " + reqdata + "}"
	reqbody := "{\"touser\":\"" + openid + "\", \"template_id\":\"" + templateid + "\", \"data\": " + reqdata + "}"

	resp, err := http.Post(url,
		"application/x-www-form-urlencoded",
		strings.NewReader(string(reqbody)))
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(body))
}
