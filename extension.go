package porthos

import "time"

type Extension interface {
	ServerListening(server Server)
	IncomingRequest(req Request)
	OutgoingResponse(req Request, res Response, resTime time.Duration, statusCode int16)
}
