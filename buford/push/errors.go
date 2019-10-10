package push

import (
	"errors"
	"fmt"
	"time"
)

// Error responses from Apple
type Error struct {
	Reason    error
	Status    int // http StatusCode
	Timestamp time.Time
}

// Service error responses.
var (
	// These could be checked prior to sending the request to Apple.
	ErrPayloadEmpty    = errors.New("PayloadEmpty")
	ErrPayloadTooLarge = errors.New("PayloadTooLarge")

	// Device token errors.
	ErrMissingDeviceToken = errors.New("MissingDeviceToken")
	ErrBadDeviceToken     = errors.New("BadDeviceToken")
	ErrTooManyRequests    = errors.New("TooManyRequests")

	// Header errors.
	ErrBadMessageID      = errors.New("BadMessageID")
	ErrBadExpirationDate = errors.New("BadExpirationDate")
	ErrBadPriority       = errors.New("BadPriority")
	ErrBadTopic          = errors.New("BadTopic")
	ErrInvalidPushType   = errors.New("InvalidPushType")

	// Certificate and topic errors.
	ErrBadCertificate            = errors.New("BadCertificate")
	ErrBadCertificateEnvironment = errors.New("BadCertificateEnvironment")
	ErrForbidden                 = errors.New("Forbidden")
	ErrMissingTopic              = errors.New("MissingTopic")
	ErrTopicDisallowed           = errors.New("TopicDisallowed")
	ErrUnregistered              = errors.New("Unregistered")
	ErrDeviceTokenNotForTopic    = errors.New("DeviceTokenNotForTopic")

	// These errors should never happen when using Push.
	ErrDuplicateHeaders = errors.New("DuplicateHeaders")
	ErrBadPath          = errors.New("BadPath")
	ErrMethodNotAllowed = errors.New("MethodNotAllowed")

	// Fatal server errors.
	ErrIdleTimeout         = errors.New("IdleTimeout")
	ErrShutdown            = errors.New("Shutdown")
	ErrInternalServerError = errors.New("InternalServerError")
	ErrServiceUnavailable  = errors.New("ServiceUnavailable")
)

// mapErrorReason converts Apple error responses into exported Err variables
// for comparisons.
func mapErrorReason(reason string) error {
	var e error
	switch reason {
	case "PayloadEmpty":
		e = ErrPayloadEmpty
	case "PayloadTooLarge":
		e = ErrPayloadTooLarge
	case "BadTopic":
		e = ErrBadTopic
	case "TopicDisallowed":
		e = ErrTopicDisallowed
	case "BadMessageId":
		e = ErrBadMessageID
	case "BadExpirationDate":
		e = ErrBadExpirationDate
	case "BadPriority":
		e = ErrBadPriority
	case "MissingDeviceToken":
		e = ErrMissingDeviceToken
	case "BadDeviceToken":
		e = ErrBadDeviceToken
	case "DeviceTokenNotForTopic":
		e = ErrDeviceTokenNotForTopic
	case "Unregistered":
		e = ErrUnregistered
	case "DuplicateHeaders":
		e = ErrDuplicateHeaders
	case "BadCertificateEnvironment":
		e = ErrBadCertificateEnvironment
	case "BadCertificate":
		e = ErrBadCertificate
	case "Forbidden":
		e = ErrForbidden
	case "BadPath":
		e = ErrBadPath
	case "MethodNotAllowed":
		e = ErrMethodNotAllowed
	case "TooManyRequests":
		e = ErrTooManyRequests
	case "IdleTimeout":
		e = ErrIdleTimeout
	case "Shutdown":
		e = ErrShutdown
	case "InternalServerError":
		e = ErrInternalServerError
	case "ServiceUnavailable":
		e = ErrServiceUnavailable
	case "MissingTopic":
		e = ErrMissingTopic
	case "InvalidPushType":
		e = ErrInvalidPushType
	default:
		e = errors.New(reason)
	}
	return e
}

func (e *Error) Error() string {
	switch e.Reason {
	case ErrPayloadEmpty:
		return "the message payload was empty"
	case ErrPayloadTooLarge:
		return "the message payload was too large"
	case ErrMissingDeviceToken:
		return "device token was not specified"
	case ErrBadDeviceToken:
		return "bad device token"
	case ErrTooManyRequests:
		return "too many requests were made consecutively to the same device token"
	case ErrBadMessageID:
		return "the ID header value is bad"
	case ErrBadExpirationDate:
		return "the Expiration header value is bad"
	case ErrBadPriority:
		return "the apns-priority value is bad"
	case ErrBadTopic:
		return "the Topic header was invalid"
	case ErrBadCertificate:
		return "the certificate was bad"
	case ErrBadCertificateEnvironment:
		return "certificate was for the wrong environment"
	case ErrForbidden:
		return "there was an error with the certificate"
	case ErrMissingTopic:
		return "the Topic header of the request was not specified and was required"
	case ErrInvalidPushType:
		return "the apns-push-type value is invalid"
	case ErrTopicDisallowed:
		return "pushing to this topic is not allowed"
	case ErrUnregistered:
		return fmt.Sprintf("device token is inactive for the specified topic (last invalid at %v)", e.Timestamp)
	case ErrDeviceTokenNotForTopic:
		return "device token does not match the specified topic"
	case ErrDuplicateHeaders:
		return "one or more headers were repeated"
	case ErrBadPath:
		return "the request contained a bad :path"
	case ErrMethodNotAllowed:
		return "the specified :method was not POST"
	case ErrIdleTimeout:
		return "idle time out"
	case ErrShutdown:
		return "the server is shutting down"
	case ErrInternalServerError:
		return "an internal server error occurred"
	case ErrServiceUnavailable:
		return "the service is unavailable"
	default:
		return fmt.Sprintf("unknown error: %v", e.Reason.Error())
	}
}
