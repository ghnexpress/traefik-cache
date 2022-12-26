package log

const (
	MiddlewareName = "middlewareName"
	MiddlewareType = "middlewareType"
)

// // GetLogger creates a logger with the middleware fields.
// func GetLogger(ctx context.Context, middleware, middlewareType string) *zerolog.Logger {
// 	logger := log.Ctx(ctx).With().Str(MiddlewareName, middleware).Str(MiddlewareType, middlewareType).Logger()

// 	return &logger
// }

func Log(text string) {
	// os.Stderr.WriteString(text)
}
