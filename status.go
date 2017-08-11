package porthos

// Our status codes were inherited from HTTP status sodes:
// See: http://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml

const (
	StatusOK                          int32 = 200
	StatusCreated                     int32 = 201
	StatusAccepted                    int32 = 202
	StatusNonAuthoritativeInfo        int32 = 203
	StatusNoContent                   int32 = 204
	StatusResetContent                int32 = 205
	StatusPartialContent              int32 = 206
	StatusMovedPermanently            int32 = 301
	StatusFound                       int32 = 302
	StatusNotModified                 int32 = 304
	StatusBadRequest                  int32 = 400
	StatusUnauthorized                int32 = 401
	StatusForbidden                   int32 = 403
	StatusNotFound                    int32 = 404
	StatusMethodNotAllowed            int32 = 405
	StatusNotAcceptable               int32 = 406
	StatusConflict                    int32 = 409
	StatusGone                        int32 = 410
	StatusLocked                      int32 = 423
	StatusFailedDependency            int32 = 424
	StatusPreconditionRequired        int32 = 428
	StatusTooManyRequests             int32 = 429
	StatusRequestHeaderFieldsTooLarge int32 = 431
	StatusUnavailableForLegalReasons  int32 = 451
	StatusInternalServerError         int32 = 500
	StatusNotImplemented              int32 = 501
	StatusServiceUnavailable          int32 = 503
	StatusInsufficientStorage         int32 = 507
)
