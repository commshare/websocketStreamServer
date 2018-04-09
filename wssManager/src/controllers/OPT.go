package controllers

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	mrand "math/rand"

	"github.com/astaxie/beego"
)

type StreamList struct {
	Status  int          `json:"status"` //200 ok
	Streams []StreamInfo `json:"streams"`
}

type StreamInfo struct {
	StreamName  string `json:"streamName"`
	Connections int    `json:"connections"`
}

const (
	WS_SHOW_ALL_STREAM = iota
	WS_GET_LIVE_PLAYER_COUNT
	WS_ENABLE_BLACK_LIST
	WS_SET_BLACK_LIST
	WS_ENABLE_WHITE_LIST
	WS_SET_WHITE_LIST
	WS_SET_UP_STREAM_APP
	WS_PULL_RTMP_STREAM
	WS_ADD_SINK
	WS_DEL_SINK
	WS_ADD_SOURCE
	WS_DEL_SOURCE
	WS_GET_SOURCE
)

type OPTController struct {
	beego.Controller
}

func (this *OPTController) Post() {
	session := this.GetSession(tokenName)
	if nil == session {
		beego.Debug("no session")
		streams := &StreamList{}
		data, _ := json.Marshal(streams)
		this.Ctx.Output.Body(data)
		return
	}
	op, err := this.GetInt("op", -1)
	if err != nil {
		beego.Debug("bad op")
		streams := &StreamList{}
		data, _ := json.Marshal(streams)
		this.Ctx.Output.Body(data)
		return
	}
	switch op {
	case WS_SHOW_ALL_STREAM:
		this.Ctx.Output.Body(this.getAllStream())
	default:
		beego.Debug("bad op")
		streams := &StreamList{}
		data, _ := json.Marshal(streams)
		this.Ctx.Output.Body(data)
	}
}

func getMd5String(str string) string {
	h := md5.New()

	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
func GenerateGUID() string {
	b := make([]byte, 48)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}

	return getMd5String(base64.URLEncoding.EncodeToString(b))
}
func (this *OPTController) getAllStream() []byte {
	streams := &StreamList{}
	count := mrand.Intn(50) + 1
	streams.Streams = make([]StreamInfo, count)
	streams.Status = 200
	for idx := 0; idx < count; idx++ {
		streams.Streams[idx].Connections = mrand.Intn(100) + 1
		streams.Streams[idx].StreamName = GenerateGUID()
	}
	out, err := json.Marshal(streams)
	if err != nil {
		beego.Debug(err.Error())
		return nil
	}
	beego.Debug(string(out))
	return out
}
