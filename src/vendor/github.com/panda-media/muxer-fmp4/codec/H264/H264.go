package H264

import (
	"bytes"
)

const (
	NAL_SLICE           = 1 //has slice header
	NAL_DPA             = 2 //has slice header
	NAL_DPB             = 3
	NAL_DPC             = 4
	NAL_IDR_SLICE       = 5 //has slice header
	NAL_SEI             = 6
	NAL_SPS             = 7
	NAL_PPS             = 8
	NAL_AUD             = 9
	NAL_END_SEQUENCE    = 10
	NAL_END_STREAM      = 11
	NAL_FILLER_DATA     = 12
	NAL_SPS_EXT         = 13
	NAL_AUXILIARY_SLICE = 19
)

func emulation_prevention(nal []byte) []byte {
	buf := bytes.Buffer{}
	for i := 0; i < len(nal); i++ {
		if i+2 < len(nal) {
			if 0 == (nal[i]^0x00)+(nal[i+1]^0x00)+(nal[i+2]^0x03) {
				buf.WriteByte(nal[i])
				i++
				buf.WriteByte(nal[i])
				i++
				continue
			}
		}
		buf.WriteByte(nal[i])
	}
	return buf.Bytes()
}

func DecodeSPS(sps_data []byte) (width, height, fps int, chroma_format_idc, bit_depth_luma_minus8, bit_depth_chroma_minus8 byte) {

	data := emulation_prevention(sps_data)

	sps_info := decodeSPS_RBSP(data[1:])
	width = sps_info.width
	height = sps_info.height
	if sps_info.vui != nil && sps_info.vui.timing_info_present_flag != 0 && sps_info.vui.time_scale != 0 && sps_info.vui.num_units_in_tick != 0 {
		fps = sps_info.vui.time_scale / (sps_info.vui.num_units_in_tick * 2)
	} else {
		fps = -1
	}

	chroma_format_idc = byte(sps_info.chroma_format_idc)
	bit_depth_chroma_minus8 = byte(sps_info.bit_depth_chroma_minus8)
	bit_depth_luma_minus8 = byte(sps_info.bit_depth_luma_minus8)

	return
}

type H264TimeCalculator struct {
	sps              *SPS
	frame_counter    int64
	frame_duration   int64
	group_size       int64
	last_frame_group int64
	last_cnt_lsb     int
	last_group_time  int64
	next_group_time  int64
	fps              int64
}

func (this *H264TimeCalculator) SetSPS(sps []byte, fps int) {
	if this.sps != nil {
		return
	}
	data := emulation_prevention(sps)
	this.sps = decodeSPS_RBSP(data[1:])
	this.frame_duration = int64(1000 / fps)
	this.fps = int64(fps)
	this.group_size = ((1 << uint(this.sps.log2_max_pic_order_cnt_lsb_minus4+4)) + 1) / 2
	this.last_frame_group = 0
	this.last_cnt_lsb = 0
	this.last_group_time = 0
	this.next_group_time = 0
}

func (this *H264TimeCalculator) getTimestamp() (timestamp int64) {
	timestamp = this.frame_counter * 1000 / this.fps
	return
}

func (this *H264TimeCalculator) AddNal(data []byte, timestamp int64) (pts, cts int64, bFrame bool) {
	nalType := data[0] & 0x1f

	if nalType == NAL_SLICE || nalType == NAL_DPA ||
		nalType == NAL_IDR_SLICE {
		bFrame = true
		if this.sps.pic_order_cnt_type == 0 {
			pts, cts = this.cnt_type_0(data)
		} else if this.sps.pic_order_cnt_type == 1 {
			pts, cts = this.cnt_type_1(data)
		} else if this.sps.pic_order_cnt_type == 2 {
			pts, cts = this.cnt_type_2(data)
		}

		if timestamp != 0 {
			pts = int64(timestamp)
		}
		this.frame_counter++
	} else {
		return 0, 0, false
	}
	return
}

func (this *H264TimeCalculator) cnt_type_0(data []byte) (pts, cts int64) {
	pts = this.getTimestamp()
	header := decodeNalSliceHeader(data, this.sps)
	lsb := header.pic_order_cnt_lsb / 2

	if lsb == 0 {
		dts := pts + this.frame_duration*2
		cts = dts - pts
		this.last_group_time = dts
	} else {
		var op int64
		if lsb < this.last_cnt_lsb && (this.last_cnt_lsb-lsb) > int(this.group_size/2) {
			//next
			op = 1
		} else if lsb > this.last_cnt_lsb && (lsb-this.last_cnt_lsb) > int(this.group_size/2) {
			//last
			op = -1
		} else {
			op = 0
		}

		switch op {
		case -1:
			this.last_group_time -= this.frame_duration * this.group_size
			dts := this.last_group_time + int64(lsb)*this.frame_duration
			cts = dts - pts
		case 0:
			dts := this.last_group_time + int64(lsb)*this.frame_duration
			cts = dts - pts
		case 1:
			this.last_group_time += this.group_size * this.frame_duration
			dts := this.last_group_time + int64(lsb)*this.frame_duration
			cts = dts - pts
		}
	}
	if cts < 0 {
		cts = 0
	}
	this.last_cnt_lsb = lsb
	return
}

func (this *H264TimeCalculator) cnt_type_1(data []byte) (pts, cts int64) {
	pts = this.getTimestamp()
	return
}

func (this *H264TimeCalculator) cnt_type_2(data []byte) (pts, cts int64) {
	pts = this.getTimestamp()
	return
}
