package handler

//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func (h *Handler) parseTaskPathID(id string) (string, error) {
	return h.httpPort.ParseBoundedPathID("taskcore.handler.parseTaskPathID", id, "id")
}
