package porthos

// Our status codes were inherited from HTTP status sodes:
// See: http://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml

const (
	StatusOK                          int16 = 200
	StatusCreated                     int16 = 201
	StatusAccepted                    int16 = 202
	StatusNonAuthoritativeInfo        int16 = 203
	StatusNoContent                   int16 = 204
	StatusResetContent                int16 = 205
	StatusPartialContent              int16 = 206
	StatusMovedPermanently            int16 = 301
	StatusFound                       int16 = 302
	StatusNotModified                 int16 = 304
	StatusBadRequest                  int16 = 400
	StatusUnauthorized                int16 = 401
	StatusForbidden                   int16 = 403
	StatusNotFound                    int16 = 404
	StatusMethodNotAllowed            int16 = 405
	StatusNotAcceptable               int16 = 406
	StatusConflict                    int16 = 409
	StatusGone                        int16 = 410
	StatusLocked                      int16 = 423
	StatusFailedDependency            int16 = 424
	StatusPreconditionRequired        int16 = 428
	StatusTooManyRequests             int16 = 429
	StatusRequestHeaderFieldsTooLarge int16 = 431
	StatusUnavailableForLegalReasons  int16 = 451
	StatusInternalServerError         int16 = 500
	StatusNotImplemented              int16 = 501
	StatusServiceUnavailable          int16 = 503
	StatusInsufficientStorage         int16 = 507
)
