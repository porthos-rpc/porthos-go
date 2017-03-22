package porthos

// Our status codes were inherited from HTTP status sodes:
// See: http://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml

const (
	StatusOK                          = 200
	StatusCreated                     = 201
	StatusAccepted                    = 202
	StatusNonAuthoritativeInfo        = 203
	StatusNoContent                   = 204
	StatusResetContent                = 205
	StatusPartialContent              = 206
	StatusMovedPermanently            = 301
	StatusFound                       = 302
	StatusNotModified                 = 304
	StatusBadRequest                  = 400
	StatusUnauthorized                = 401
	StatusForbidden                   = 403
	StatusNotFound                    = 404
	StatusMethodNotAllowed            = 405
	StatusNotAcceptable               = 406
	StatusConflict                    = 409
	StatusGone                        = 410
	StatusLocked                      = 423
	StatusFailedDependency            = 424
	StatusPreconditionRequired        = 428
	StatusTooManyRequests             = 429
	StatusRequestHeaderFieldsTooLarge = 431
	StatusUnavailableForLegalReasons  = 451
	StatusInternalServerError         = 500
	StatusNotImplemented              = 501
	StatusServiceUnavailable          = 503
	StatusInsufficientStorage         = 507
)
