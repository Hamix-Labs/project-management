package handler

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

const maxTaskEventSeqParamBytes = 32

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parseTaskPathID(id string) (string, error) {
	return handlerhttp.ParseBoundedPathID("taskevents.handler.parseTaskPathID", id, "id")
}
