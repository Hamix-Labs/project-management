package handler

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parsePathID(id string) (string, error) {
	return handlerhttp.ParsePathID(id)
}
