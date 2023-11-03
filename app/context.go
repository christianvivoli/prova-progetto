package app

import (
	"context"
)

// contextKey represents an internal key for adding context fields.
// This is considered best practice as it prevents other packages from
// interfering with our context keys.
type contextKey int

const (
	// stores the user logged in
	userContextKey = contextKey(iota + 1)
	roleContextKey
	txContextKey
	HttpRequestTypeKey
	localeContextKey
	deviceContextKey

	ContextParamClaims = "claims"
	// ContextParamRole            = "role"
	ContextParamUser            = "user"
	ContextParamHttpRequestType = "req_type"

	HttpRequestTypeAdmin = "admin"
	HttpRequestTypeAPI   = "api"
)

// Background returns a new background context with the root role attached.
func Background() context.Context {
	return context.Background()
}

// // Now returns current now time stored in the context, otherwise is used time.Now().UTC()
// func Now(ctx context.Context) time.Time {
// 	if tx := TxFromContext(ctx); tx != nil {
// 		return tx.Now()
// 	}
// 	return time.Now().UTC()
// }

// // NewContextWithLocale returns a new context with the provided locale attached.
// func NewContextWithLocale(ctx context.Context, locale string) context.Context {
// 	return context.WithValue(ctx, localeContextKey, locale)
// }

// // NewContextWithTx returns a new context with provided tx attached.
// // This ca be useful to implements multi layer transactions.
// // func NewContextWithTx(ctx context.Context, tx PendingTx) context.Context {
// // 	return context.WithValue(ctx, txContextKey, tx)
// // }

// // NewContextWithUser returns a new context with the provided user attached.
// func NewContextWithUser(ctx context.Context, user *User) context.Context {
// 	return context.WithValue(ctx, userContextKey, user)
// }

// // NewContextWithHttpRequestType returns a new context with the previded http req type attached.
// func NewContextWithHttpRequestType(ctx context.Context, reqType string) context.Context {
// 	return context.WithValue(ctx, HttpRequestTypeKey, reqType)
// }

// // HttpRequestTypeFromContext returns req type stored in the provided context.
// func HttpRequestTypeFromContext(ctx context.Context) string {
// 	if v, ok := ctx.Value(HttpRequestTypeKey).(string); ok {
// 		return v
// 	}
// 	return ""
// }

// // IsHttpRequestTypeAdmin returns true if the provided context has a http req type admin attached.
// func IsHttpRequestTypeAdmin(ctx context.Context) bool {
// 	return HttpRequestTypeFromContext(ctx) == HttpRequestTypeAdmin
// }

// // IsHttpRequestTypeAPI returns true if the provided context has a http req type api attached.
// func IsHttpRequestTypeAPI(ctx context.Context) bool {
// 	return HttpRequestTypeFromContext(ctx) == HttpRequestTypeAPI
// }

// // IsLocalizedContext returns true if the provided context has a locale attached.
// func IsLocalizedContext(ctx context.Context) bool {
// 	return rawLocaleFromContext(ctx) != ""
// }

// // LocaleFromContext returns the locale stored in the provided context, if no locale is stored, the default locale is returned.
// func LocaleFromContext(ctx context.Context) string {
// 	locale := rawLocaleFromContext(ctx)
// 	if locale == "" {
// 		return DefaultLocale
// 	}
// 	return locale
// }

// // RawLocaleFromContext returns the raw locale stored in the provided context.
// func rawLocaleFromContext(ctx context.Context) string {
// 	if ctx == nil {
// 		return ""
// 	}
// 	locale, ok := ctx.Value(localeContextKey).(string)
// 	if !ok {
// 		return ""
// 	}
// 	return locale
// }

// TxFromContext returns the transaction stored inside the context.
func TxFromContext(ctx context.Context) PendingTx {
	if ctx == nil {
		return nil
	}
	tx, ok := ctx.Value(txContextKey).(PendingTx)
	if !ok {
		return nil
	}
	return tx
}

// // UserFromContext returns the user stored in the provided context.
// func UserFromContext(ctx context.Context) *User {
// 	if ctx == nil {
// 		return nil
// 	}
// 	user, ok := ctx.Value(userContextKey).(*User)
// 	if !ok {
// 		return nil
// 	}
// 	return user
// }

// // UserIDFromContext returns the user ID stored in the provided context.
// func UserIDFromContext(ctx context.Context) int64 {
// 	if user := UserFromContext(ctx); user != nil {
// 		return user.ID
// 	}
// 	return 0
// }
