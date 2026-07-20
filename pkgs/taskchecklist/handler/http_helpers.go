package handler

import (
	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/handlerhttp"
)

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parseTaskPathID(id string) (string, error) {
	return handlerhttp.ParseBoundedPathID("taskchecklist.handler.parseTaskPathID", id, "id")
}

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func parseTaskPathItemID(id string) (string, error) {
	return handlerhttp.ParseBoundedPathID("taskchecklist.handler.parseTaskPathItemID", id, "item id")
}
