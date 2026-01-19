package graphql

import "context"

type key string

const (
	initpayload key = "ws_initpayload_context"
	closeReason key = "ws_close_reason_context"
)

// InitPayload is a structure that is parsed from the websocket init message payload. TO use
// request headers for non-websocket, instead wrap the graphql handler in a middleware.
type InitPayload map[string]any

// GetString safely gets a string value from the payload. It returns an empty string if the
// payload is nil or the value isn't set.
func (p InitPayload) GetString(key string) string {
	if p == nil {
		return ""
	}

	if value, ok := p[key]; ok {
		res, _ := value.(string)
		return res
	}

	return ""
}

// Authorization is a short hand for getting the Authorization header from the
// payload.
func (p InitPayload) Authorization() string {
	if value := p.GetString("Authorization"); value != "" {
		return value
	}

	if value := p.GetString("authorization"); value != "" {
		return value
	}

	return ""
}

func withInitPayload(ctx context.Context, payload InitPayload) context.Context {
	return context.WithValue(ctx, initpayload, payload)
}

// GetInitPayload gets a map of the data sent with the connection_init message, which is used by
// graphql clients as a stand-in for HTTP headers.
func GetInitPayload(ctx context.Context) InitPayload {
	payload, ok := ctx.Value(initpayload).(InitPayload)
	if !ok {
		return nil
	}

	return payload
}

// withCloseReason adds a close reason to the context that will be sent to the client
// when the connection is closed.
func withCloseReason(ctx context.Context, reason string) context.Context {
	return context.WithValue(ctx, closeReason, reason)
}

// closeReasonForContext retrieves the close reason from the context, if any.
func closeReasonForContext(ctx context.Context) string {
	reason, _ := ctx.Value(closeReason).(string)
	return reason
}
