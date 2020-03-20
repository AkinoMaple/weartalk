package weartalk

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/valyala/fasthttp"
)

type WearTalk struct {
	UID      int
	NickName string
	Sex      int8
	Key      string
	Device   int8
	XFF      string
}

type Version struct {
	Version string `json:"version"`
	Tips    string `json:"tips"`
	URL     string `json:"url"`
}

type UserInfo struct {
	// shit English suki
	Phone   []uint
	Name    []string
	email   string
	message int
}

func (wt *WearTalk) getNickName() string {
	switch wt.Sex {
	case 0:
		return wt.NickName + "♂"

	case 1:
		return wt.NickName + "♀"

	default:
		return wt.NickName
	}
}

func (wt *WearTalk) getUID() string {
	switch wt.Device {
	case 0:
		return "TaMi" + strconv.Itoa(wt.UID)
	case 1:
		return "TaMP" + strconv.Itoa(wt.UID)
	default:
		return "Ta" + strconv.Itoa(wt.UID)
	}
}

func (wt *WearTalk) Get(url string, args *fasthttp.Args) ([]byte, error) {
	req := &fasthttp.Request{}
	req.SetRequestURI(url + "?" + args.String())
	req.Header.SetMethod("GET")
	req.Header.SetUserAgent("Dalvik/2.1.0 (Linux; U; Android 5.1.1; PRO 6 Plus Build/LMY48Z)")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip")
	if wt.XFF != "" {
		req.Header.Set("X-Forwarded-For", wt.XFF)
	}

	resp := &fasthttp.Response{}
	client := &fasthttp.Client{}

	err := client.Do(req, resp)

	return resp.Body(), err
}

func (wt *WearTalk) GetVersion() (*Version, error) {
	args := &fasthttp.Args{}
	args.Add("ruanjian", "weartalk参数Ta")
	resp, rErr := wt.Get("https://zhinengjiaju.vip/xczx/getversion.action", args)
	if rErr != nil {
		log.Printf("Get Version Request Error: %s\n", rErr)
		return nil, rErr
	}

	var ver Version
	mErr := json.Unmarshal(resp, &ver)
	if mErr != nil {
		log.Printf("Unmarshal Version Data Error: %s\n", mErr)
		return nil, mErr
	}

	return &ver, nil
}

func (wt *WearTalk) Send(roomid string, message string, arguments ...int64) (map[string]string, error) {
	var timestamp int64
	if arguments != nil {
		timestamp = arguments[0]
	}

	args := &fasthttp.Args{}
	args.Add("roomid", roomid)
	args.Add("uid", wt.getUID())
	args.Add("words", message)
	args.Add("nickname", wt.getNickName())
	if timestamp == 0 {
		timestamp = time.Now().Unix()
	}
	args.Add("timestamp", strconv.FormatInt(timestamp, 10))

	if wt.Key == "" {
		wt.Key = "fnakdjsgangaj65984qdvcvo71as1a3ds1g56a1g5a1ggagra&gajg15615avasggsa66a15g651a71ger1g5ar1g56ytiu7"
	}
	args.Add("key", wt.Key)

	salt, sErr := wt.caclSalt(args)
	if sErr != nil {
		log.Printf("Calculate Salt Error: %s\n", sErr)
		return nil, sErr
	}
	args.Add("salt", salt)

	args.Del("key")

	resp, rErr := wt.Get("http://zhinengjiaju.vip/xczx/saidwords.action", args)
	if rErr != nil {
		log.Printf("Send Request Error: %s\n", rErr)
		return nil, rErr
	}

	if resp == nil {
		return nil, errors.New("Send Response is nil, Maybe IP was be blocked.")
	}

	status := make(map[string]string)
	mErr := json.Unmarshal(resp, &status)
	if mErr != nil {
		log.Printf("Unmarshal Sended Status Data Error: %s\n", mErr)
		return nil, mErr
	}

	return status, nil
}

func (wt *WearTalk) caclSalt(args *fasthttp.Args) (string, error) {
	cipher := md5.New()
	_, err := cipher.Write(args.QueryString())
	return hex.EncodeToString(cipher.Sum(nil)), err
}
