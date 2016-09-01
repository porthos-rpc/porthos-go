package status

// Our status codes were inherited from HTTP status sodes:
// See: http://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml

const (
	OK                   = 200
	Created              = 201
	Accepted             = 202
	NonAuthoritativeInfo = 203
	NoContent            = 204
	ResetContent         = 205
	PartialContent       = 206

	MovedPermanently = 301
	Found            = 302
	NotModified      = 304

	BadRequest                  = 400
	Unauthorized                = 401
	Forbidden                   = 403
	NotFound                    = 404
	MethodNotAllowed            = 405
	NotAcceptable               = 406
	Conflict                    = 409
	Gone                        = 410
	Locked                      = 423
	FailedDependency            = 424
	PreconditionRequired        = 428
	TooManyRequests             = 429
	RequestHeaderFieldsTooLarge = 431
	UnavailableForLegalReasons  = 451

	InternalServerError = 500
	NotImplemented      = 501
	ServiceUnavailable  = 503
	InsufficientStorage = 507
)
