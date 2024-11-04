package types

type ErrorCode int

// error code
const (
	ErrCodeParam ErrorCode = iota + 1001
	ErrCodeParse
	ErrCodeJson
	ErrCodeProtobuf
	ErrCodeTimeout
	ErrCodeRendezvous
	ErrCodeModel
	ErrCodeUpload
	ErrCodeBuffer
	ErrCodePermission
	ErrCodeUnsupported
	ErrCodeHostInfo
	ErrCodeEncrypt
	ErrCodeDecrypt
	ErrCodeUUID
	ErrCodeDatabase
	ErrCodeProxy
	ErrCodeStream
	ErrCodeDeprecated
	ErrCodeInternal ErrorCode = 5000
)

// error message
var errMsg = map[ErrorCode]string{
	ErrCodeParam:       "Parameter error",
	ErrCodeParse:       "Parsing error",
	ErrCodeJson:        "Json error",
	ErrCodeProtobuf:    "Protobuf serialization error",
	ErrCodeTimeout:     "Processing timeout",
	ErrCodeRendezvous:  "Rendezvous error",
	ErrCodeModel:       "Model error",
	ErrCodeUpload:      "Upload error",
	ErrCodeBuffer:      "Buffer error",
	ErrCodePermission:  "Permission error",
	ErrCodeUnsupported: "Unsupported function",
	ErrCodeHostInfo:    "Host info error",
	ErrCodeEncrypt:     "Encrypt error",
	ErrCodeDecrypt:     "Decrypt error",
	ErrCodeUUID:        "UUID error",
	ErrCodeDatabase:    "Database error",
	ErrCodeProxy:       "Proxy error",
	ErrCodeStream:      "Stream error",
	ErrCodeDeprecated:  "Deprecated function",
	ErrCodeInternal:    "Internal server error",
}

func (e ErrorCode) String() string {
	return errMsg[e]
}
