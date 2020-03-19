package main

import (
	"encoding/json"
	"log"
	"strconv"

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

func (wt *WearTalk) GetVersion() *Version {
	args := &fasthttp.Args{}
	args.Add("ruanjian", "weartalk参数Ta")
	resp, rErr := wt.Get("https://zhinengjiaju.vip/xczx/getversion.action", args)
	if rErr != nil {
		log.Printf("Get Version Request Error: %s", rErr)
		return nil
	}

	var ver Version
	mErr := json.Unmarshal(resp, &ver)
	if mErr != nil {
		log.Printf("Unmarshal Version Data Error: %s", mErr)
	}

	return &ver
}
