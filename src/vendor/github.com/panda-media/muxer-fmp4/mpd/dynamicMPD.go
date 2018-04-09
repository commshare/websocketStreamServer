package mpd

import (
	"container/list"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/panda-media/muxer-fmp4/utils"
	"strconv"
	"sync"
	"time"
)

const (
	IdAudio        = "audio0"
	IdVideo        = "video0"
	MillInSec      = 1000
	MillInSecFloat = 1000.0
)

type videoInfo struct {
	timeScale int
	width     int
	height    int
	frameRate int
	bitrate   int
	codecs    string
}
type audioInfo struct {
	timescale  int
	sampleRate int
	bandwidth  int
	channel    int
	sampelSize int
	codecs     string
}

type segmentTimeData struct {
	t    int64
	d    int
	data []byte
}

type MPDDynamic struct {
	minBufferMS        int
	maxSliceCount      int
	vide               *videoInfo
	audi               *audioInfo
	videoData          map[int64]*segmentTimeData
	videoKeys          *list.List
	lastVideoTimestamp int64
	muxVideo           sync.RWMutex
	audioData          map[int64]*segmentTimeData
	audioKeys          *list.List
	lastAudioTimestamp int64
	muxAudio           sync.RWMutex
	mediaDataStartTime time.Time
}

func NewDynamicMPDCreater(minBufferDurationMS, maxSliceCount int) (mpd *MPDDynamic) {
	mpd = &MPDDynamic{}
	mpd.init(minBufferDurationMS, maxSliceCount)
	return
}

func (this *MPDDynamic) init(minBufferDurationMS, maxSliceCount int) {
	this.minBufferMS = minBufferDurationMS
	this.maxSliceCount = maxSliceCount
	this.videoKeys = list.New()
	this.audioKeys = list.New()
	this.videoData = make(map[int64]*segmentTimeData)
	this.audioData = make(map[int64]*segmentTimeData)
}
func (this *MPDDynamic) generatePTime(year, month, day, hour, minute, sec, mill int) (pt string) {
	pt = "P"
	if year > 0 {
		pt += strconv.Itoa(year) + "Y"
	}
	if month > 0 {
		pt += strconv.Itoa(month) + "M"
	}
	if day > 0 {
		pt += strconv.Itoa(day) + "D"
	}
	pt += "T"
	if hour > 0 {
		pt += strconv.Itoa(hour) + "H"
	}
	if minute > 0 {
		pt += strconv.Itoa(minute) + "M"
	}
	pt += fmt.Sprintf("%.3f", float32(float32(sec)+float32(mill)/MillInSecFloat)) + "S"
	return
}

func (this *MPDDynamic) generatePTimeMillSec(mill int) (pt string) {
	pt = "PT" + fmt.Sprintf("%.3f", float32(float32(mill)/MillInSecFloat)) + "S"
	return
}

func (this *MPDDynamic) SetVideoInfo(timescale, width, height, frameRate, bandwidth int, codecs string) (err error) {
	if this.vide != nil {
		return errors.New("set video info for mpd again")
	}
	this.vide = &videoInfo{timescale,
		width,
		height,
		frameRate,
		bandwidth,
		codecs}
	this.mediaDataStartTime = time.Now().UTC()
	return
}

func (this *MPDDynamic) SetAudioInfo(timescale, sampleRate, bandwidth, channel, sampleSize int, codecs string) (err error) {
	if this.audi != nil {
		return errors.New("set audio info in mpd times")
	}
	this.audi = &audioInfo{timescale,
		sampleRate,
		bandwidth,
		channel,
		sampleSize,
		codecs}
	return
}

func (this *MPDDynamic) SetVideoBitrate(bitrate int) {
	this.vide.bitrate = bitrate
}
func (this *MPDDynamic) SetAudioBitrate(bitrate int) {
	this.audi.bandwidth = bitrate
}

func (this *MPDDynamic) AddVideoSlice(duration int, data []byte) (lastTimestamp int64, err error) {
	if nil == this.vide {
		err = errors.New("video info not seted")
		return
	}
	if duration<=0{
		err=errors.New("invalid video segment data")
		return
	}
	this.muxVideo.Lock()
	defer this.muxVideo.Unlock()
	segment_time_data := &segmentTimeData{}
	segment_time_data.t = this.lastVideoTimestamp
	segment_time_data.d = duration
	segment_time_data.data = data
	this.videoData[this.lastVideoTimestamp] = segment_time_data
	this.videoKeys.PushBack(this.lastVideoTimestamp)

	lastTimestamp = this.lastVideoTimestamp
	this.lastVideoTimestamp += int64(segment_time_data.d)
	if this.videoKeys.Len() > this.maxSliceCount {
		k := this.videoKeys.Front().Value.(int64)
		this.videoKeys.Remove(this.videoKeys.Front())
		delete(this.videoData, k)
	}
	return
}

func (this *MPDDynamic) AddAudioSlice(duration int, data []byte) (lastTimestamp int64, err error) {
	if nil == this.audi {
		err = errors.New("audio info not seted")
		return
	}
	if duration<=0{
		err=errors.New("invalid audio segment data")
		return
	}
	this.muxAudio.Lock()
	defer this.muxAudio.Unlock()
	segment_time_data := &segmentTimeData{}
	segment_time_data.t = this.lastAudioTimestamp
	segment_time_data.d = duration
	segment_time_data.data = data
	this.audioData[this.lastAudioTimestamp] = segment_time_data
	this.audioKeys.PushBack(this.lastAudioTimestamp)

	lastTimestamp = this.lastAudioTimestamp

	this.lastAudioTimestamp += int64(segment_time_data.d)
	if this.audioKeys.Len() > this.maxSliceCount {
		k := this.audioKeys.Front().Value.(int64)
		this.audioKeys.Remove(this.audioKeys.Front())
		delete(this.audioData, k)
	}
	return
}

func (this *MPDDynamic) GetVideoSlice(timestamp int64) (data []byte, err error) {
	this.muxVideo.RLock()
	defer this.muxVideo.RUnlock()
	seg_time_data, ok := this.videoData[timestamp]
	if false == ok {
		err = errors.New("video slice not found")
		return
	}
	data = seg_time_data.data
	return
}

func (this *MPDDynamic) GetAudioSlice(timestamp int64) (data []byte, err error) {
	this.muxAudio.RLock()
	defer this.muxAudio.RUnlock()
	seg_time_data, ok := this.audioData[timestamp]
	if false == ok {
		err = errors.New("audio slice not found")
		return
	}
	data = seg_time_data.data
	return
}

func (this *MPDDynamic) GetMPDXML() (data []byte, err error) {
	if this.vide == nil && this.audi == nil {
		return nil, errors.New("not inited")
	}
	mpd := &MPD{}
	this.muxVideo.RLock()
	defer this.muxVideo.RUnlock()
	this.muxAudio.RLock()
	defer this.muxAudio.RUnlock()
	this.mpdAttrs(mpd)
	mpd.Period = make([]PeriodXML, 1)
	mpd.Period[0].Id = "0"
	mpd.Period[0].Start = "PT0.0S"

	mpd.Period[0].AdaptationSet = make([]AdaptationSetXML, 0)
	if this.vide != nil {
		this.adaptationSetVideo(&mpd.Period[0])
	}
	if this.audi != nil {
		this.adaptationSetAudio(&mpd.Period[0])
	}

	body, err := xml.Marshal(mpd)
	if err == nil {
		header := `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
		body = utils.FormatXML(body)
		data = make([]byte, len(body)+len(header))
		copy(data, []byte(header))
		copy(data[len([]byte(header)):], body)
	}
	return
}

func (this *MPDDynamic) mpdAttrs(mpd *MPD) {
	mpd.Xmlns = MPDXMLNS
	mpd.Profiles = ProfileISOLive
	mpd.Type = dynamicMPD
	mpd.Xmlns_xlink = "http://www.w3.org/1999/xlink"
	mpd.Xmlns_xsi = "http://www.w3.org/2001/XMLSchema-instance"
	mpd.Xsi_schemaLocation = "urn:mpeg:DASH:schema:MPD:2011 http://standards.iso.org/ittf/PubliclyAvailableStandards/MPEG-DASH_schema_files/DASH-MPD.xsd"
	timestamp := time.Now().UTC()

	mpd.PublishTime = timestamp.Format("2006-01-02T15:04:05.000Z")

	mpd.AvailabilityStartTime = func() (availablityStartTime string) {
		availablityStartTime = this.mediaDataStartTime.Format("2006-01-02T15:04:05.000Z")
		return
	}()

	if this.videoKeys.Len()!=0{
		this.videoBufferTime(mpd)
	}else if this.audioKeys.Len()!=0{
		this.audioBufferTime(mpd)
	}
}

func (this *MPDDynamic)videoBufferTime(mpd *MPD){

	minBufferTime := 0xfffffff
	for _, e := range this.videoData {
		if minBufferTime > e.d {
			minBufferTime = e.d
		}
	}
	minBufferTime = minBufferTime * MillInSec / this.vide.timeScale

	if this.minBufferMS<minBufferTime/3{
		mpd.MinimumUpdatePeriod=this.generatePTimeMillSec(minBufferTime/2)
	}else{
		mpd.MinimumUpdatePeriod=this.generatePTimeMillSec(this.minBufferMS)
	}
	mpd.MinBufferTime = this.generatePTimeMillSec(minBufferTime)

	mpd.SuggestedPresentationDelay = func() (suggestedPresentationDelay string) {
		delayCounts := this.maxSliceCount / 3
		delay := 0
		if this.videoKeys.Len() > delayCounts {
			e := this.videoKeys.Front()
			for i := 0; i < delayCounts; i++ {
				delay += this.videoData[e.Value.(int64)].d
			}
		}
		delay = 0
		suggestedPresentationDelay = this.generatePTimeMillSec(delay * MillInSec / this.vide.timeScale)
		return
	}()
}

func (this *MPDDynamic)audioBufferTime(mpd *MPD){
	minBufferTime := 0xfffffff
	for _, e := range this.audioData {
		if minBufferTime > e.d {
			minBufferTime = e.d
		}
	}
	minBufferTime = minBufferTime * MillInSec / this.audi.timescale

	if this.minBufferMS<minBufferTime/3{
		mpd.MinimumUpdatePeriod=this.generatePTimeMillSec(minBufferTime/2)
	}else{
		mpd.MinimumUpdatePeriod=this.generatePTimeMillSec(this.minBufferMS)
	}
	mpd.MinBufferTime = this.generatePTimeMillSec(minBufferTime)

	mpd.SuggestedPresentationDelay=this.generatePTimeMillSec(0)
}

func (this *MPDDynamic) adaptationSetVideo(period *PeriodXML) {
	ada := AdaptationSetXML{}
	ada.ContentType = "video"
	ada.Id = "0"
	ada.MimeType = "video/mp4"
	//ada.Codecs=this.vide.codecs
	ada.Width = this.vide.width
	ada.Height = this.vide.height
	ada.FrameRate = this.vide.frameRate
	ada.SegmentAlignment = true
	ada.StartWithSAP = 1
	ada.SubsegmentAlignment = true
	ada.SubsegmentStartsWithSAP = 1

	role := &RoleXML{}
	role.SchemeIdUri = "urn:mpeg:dash:role:2011"
	role.Value = "main"
	ada.Role = role

	ada.Representation = make([]RepresentationXML, 0)
	representation := RepresentationXML{}
	representation.Bandwidth = this.vide.bitrate
	representation.Codecs = this.vide.codecs
	representation.Id = IdVideo
	ada.Representation = append(ada.Representation, representation)

	ada.SegmentTemplate.TimeScale = this.vide.timeScale
	ada.SegmentTemplate.Media = "video_$RepresentationID$_$Time$_mp4.m4s"
	ada.SegmentTemplate.Initialization = "video_$RepresentationID$_init_mp4.m4s"
	segmentTimeLine := &SegmentTimelineXML{}
	segmentTimeLine.S = make([]SegmentTimelineDesc, this.videoKeys.Len())
	for e, idx := this.videoKeys.Front(), 0; e != nil; e = e.Next() {
		k := e.Value.(int64)
		segmentTimeLine.S[idx].T = int(this.videoData[k].t & 0xffffffff)
		segmentTimeLine.S[idx].D = this.videoData[k].d
		idx++
	}
	ada.SegmentTemplate.SegmentTimeline = segmentTimeLine

	period.AdaptationSet = append(period.AdaptationSet, ada)
}

func (this *MPDDynamic) adaptationSetAudio(period *PeriodXML) {
	ada := AdaptationSetXML{}
	ada.ContentType = "audio"
	ada.Id = "1"
	ada.MimeType = "audio/mp4"
	//ada.Codecs=this.audi.codecs
	ada.Lang = "eng"
	ada.SegmentAlignment = true
	ada.StartWithSAP = 1
	ada.SubsegmentAlignment = true
	ada.SubsegmentStartsWithSAP = 1

	ada.Representation = make([]RepresentationXML, 0)
	representation := RepresentationXML{}
	representation.Bandwidth = this.audi.bandwidth
	representation.Id = IdAudio
	representation.Codecs = this.audi.codecs
	representation.AudioSamplingRate = this.audi.sampleRate
	ada.Representation = append(ada.Representation, representation)

	audioChannelConfiguration := &AudioChannelConfigurationXML{}
	audioChannelConfiguration.Value = this.audi.channel
	audioChannelConfiguration.SchemeIdUri = SchemeIdUri
	ada.AudioChannelConfiguration = audioChannelConfiguration

	role := &RoleXML{}
	role.SchemeIdUri = "urn:mpeg:dash:role:2011"
	role.Value = "main"
	ada.Role = role

	ada.SegmentTemplate.TimeScale = this.audi.sampleRate
	ada.SegmentTemplate.Media = "audio_$RepresentationID$_$Time$_mp4.m4s"
	ada.SegmentTemplate.Initialization = "audio_$RepresentationID$_init_mp4.m4s"
	segmentTimeLine := &SegmentTimelineXML{}
	segmentTimeLine.S = make([]SegmentTimelineDesc, this.audioKeys.Len())
	for e, idx := this.audioKeys.Front(), 0; e != nil; e = e.Next() {
		k := e.Value.(int64)
		v := this.audioData[k]
		if nil == v {
			break
		}
		segmentTimeLine.S[idx].T = int(v.t & 0xffffffff)
		segmentTimeLine.S[idx].D = v.d
		idx++
	}
	ada.SegmentTemplate.SegmentTimeline = segmentTimeLine

	period.AdaptationSet = append(period.AdaptationSet, ada)
}
