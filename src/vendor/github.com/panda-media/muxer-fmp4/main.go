package main

import (
	//"github.com/panda-media/muxer-fmp4/example/dash"
	"log"
	"github.com/panda-media/muxer-fmp4/example/ws_fmp4"
)

func main() {
	log.SetFlags(log.Lshortfile)
	//dash.FlvFileToFMP4("111.flv")
	//ws_fmp4.RunWS_server(8080,"/ws/")
	//ws_fmp4.SaveSegment("9",3000)
	ws_fmp4.WSFMP4Demo()
	//ch:=make(chan int)
	//<-ch
	return
}
