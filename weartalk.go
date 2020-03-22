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
	UID      string //Ta/TaMP/TaMi + Int
	ID       int    // Int ID
	NickName string
	Sex      int8
	Key      string
	Device   int8
	Avatar   string
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
	Email   string
	Message int
}

type Messages struct {
	Time   interface{} `json:"time"`
	Room   Room        `json:"room"`
	Status interface{} `json:"status"`
}

type Room struct {
	Firstman string      `json:"fristman"`
	Password string      `json:"pwd"`
	Talks    interface{} `json:"talks"`
	RoomID   string      `json:"roomid"`
}

type Msg struct {
	UID      string `json:"uid"`
	IP       string `json:"ip"`
	Words    string `json:"words"`
	NickName string `json:"nickname"`
	Time     int64  `json:"time"`
	Avatar   string `json:"touxiangname"`
	RoomID   string `json:"-"`
}

type CallBackFunc func(*Msg)

func (wt *WearTalk) getNickName() string {
	switch wt.Sex {
	case 1:
		return wt.NickName + "♂"

	case 2:
		return wt.NickName + "♀"

	default:
		return wt.NickName
	}
}

func (wt *WearTalk) MarshalUID() {
	switch wt.Device {
	case 1:
		wt.UID = "TaMi" + strconv.Itoa(wt.ID)
	case 2:
		wt.UID = "TaMP" + strconv.Itoa(wt.ID)
	default:
		wt.UID = "Ta" + strconv.Itoa(wt.ID)
	}
}

func (wt *WearTalk) get(url string, args *fasthttp.Args) ([]byte, error) {
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
	resp, rErr := wt.get("https://zhinengjiaju.vip/xczx/getversion.action", args)
	if rErr != nil {
		log.Printf("Get Version Request Error: %s\n", rErr)
		return nil, rErr
	}

	var ver Version
	if mErr := json.Unmarshal(resp, &ver); mErr != nil {
		log.Printf("Unmarshal Version Data Error: %s\n", mErr)
		return nil, mErr
	}

	return &ver, nil
}

func (wt *WearTalk) Send(roomid string, message string, arguments ...int64) (map[string]string, error) {
	var timestamp int64
	if arguments != nil {
		timestamp = arguments[0]
	} else {
		timestamp = time.Now().UnixNano() / 1e6
	}

	args := &fasthttp.Args{}
	args.Add("roomid", roomid)
	args.Add("uid", wt.UID)
	args.Add("words", message)
	args.Add("nickname", wt.getNickName())
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
	args.Add("touxiangname", wt.Avatar)

	resp, rErr := wt.get("http://zhinengjiaju.vip/xczx/saidwords.action", args)
	if rErr != nil {
		log.Printf("Send Request Error: %s\n", rErr)
		return nil, rErr
	}

	if resp == nil {
		return nil, errors.New("Send Response is nil, Maybe IP was be blocked.")
	}

	status := make(map[string]string)
	if mErr := json.Unmarshal(resp, &status); mErr != nil {
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

func (wt *WearTalk) GetMessages(roomid string, timestamp int64) (*Messages, error) {
	args := &fasthttp.Args{}
	args.Add("roomid", roomid)
	args.Add("time", strconv.FormatInt(timestamp, 10))

	resp, rErr := wt.get("https://zhinengjiaju.vip/xczx/gettalks.action", args)
	if rErr != nil {
		log.Printf("Get Messages Error: %s\n", rErr)
		return nil, rErr
	}
	var talksRaw json.RawMessage
	msgs := Messages{
		Room: Room{
			Talks: &talksRaw,
		},
	}

	if mErr := json.Unmarshal(resp, &msgs); mErr != nil {
		log.Printf("Unmarshal Messages Error: %s\n", mErr)
		return nil, mErr
	}

	if msgs.Time != nil {
		var fErr error
		msgs.Time, fErr = strconv.Atoi(msgs.Time.(string))
		if fErr != nil {
			log.Printf("Time Atoi Error: %s\n", fErr)
		}
	}

	switch msgs.Status.(string) {
	case "null":
		msgs.Status = 0
	case "has news":
		msgs.Status = 1
	case "no news":
		msgs.Status = 2
	}

	if msgs.Status.(int) == 1 {
		var talks []Msg
		if mErr := json.Unmarshal(talksRaw, &talks); mErr != nil {
			log.Printf("Unmarshal Messages Error: %s\n", mErr)
			return nil, mErr
		}
		msgs.Room.Talks = talks
	}

	return &msgs, nil
}

func (wt *WearTalk) HandleMsg(roomid string, callback CallBackFunc, arguments ...int64) {
	var tick int64
	var timestamp int64

	if len(arguments) > 0 {
		tick = arguments[0]
	} else {
		tick = 0
	}

	if len(arguments) > 1 {
		timestamp = arguments[1]
	} else {
		timestamp = time.Now().UnixNano() / 1e6
	}

	go func() {
		ticker := time.Tick(time.Second * time.Duration(tick))
		for {
			if msgs, mErr := wt.GetMessages(roomid, timestamp); mErr == nil && msgs.Status.(int) == 1 {
				for _, msg := range msgs.Room.Talks.([]Msg) {
					msg.RoomID = roomid
					timestamp = time.Now().UnixNano() / 1e6
					callback(&msg)
				}
			}
			<-ticker
		}
	}()

}

func (wt *WearTalk) GetIsVIP(uid string) (bool, error) {
	args := &fasthttp.Args{}
	args.Add("username", uid)

	resp, rErr := wt.get("https://zhinengjiaju.vip/gpsfly0/getvip.action", args)
	if rErr != nil {
		log.Printf("Get VIP Status Error: %s\n", rErr)
		return false, rErr
	}

	isvip := make(map[string]string)
	if mErr := json.Unmarshal(resp, &isvip); mErr != nil {
		log.Printf("Unmarshal Sended Status Data Error: %s\n", mErr)
		return false, mErr
	}

	if isvip["isvip"] == "v" {
		return true, nil
	} else {
		return false, nil
	}

}
