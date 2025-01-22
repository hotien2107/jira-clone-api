package logging

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"jira-clone-api/common/configure"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/rs/zerolog"
)

var logger *zerolog.Logger

func InitLogger() {
	logger = createLogger()
	if configure.GetConfig().Debug {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func GetLogger() *zerolog.Logger {
	if logger == nil {
		InitLogger()
	}
	return logger
}

func createLogger() *zerolog.Logger {
	log := zerolog.New(getLogWriter()).With().Timestamp().Logger()
	return &log
}

func getLogWriter() zerolog.ConsoleWriter {
	return zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
}

// Logger variables
const (
	TagPid               = "pid"
	TagTime              = "time"
	TagReferer           = "referer"
	TagProtocol          = "protocol"
	TagID                = "id"
	TagIP                = "ip"
	TagIPs               = "ips"
	TagHost              = "host"
	TagMethod            = "method"
	TagPath              = "path"
	TagURL               = "url"
	TagUA                = "ua"
	TagLatency           = "latency"
	TagStatus            = "status"
	TagResBody           = "resBody"
	TagQueryStringParams = "queryParams"
	TagBody              = "body"
	TagBytesSent         = "bytesSent"
	TagBytesReceived     = "bytesReceived"
	TagRoute             = "route"
	TagError             = "error"
	TagHeader            = "header:"
	TagLocals            = "locals:"
	TagQuery             = "query:"
	TagForm              = "form:"
	TagCookie            = "cookie:"
)

const timeIntervalDefault = 500

type LoggerMiddlewareConfig struct {
	Next             func(ctx *fiber.Ctx) bool
	timeZoneLocation *time.Location
	TimeFormat       string
	TimeZone         string
	logger           zerolog.Logger
	Format           []string
	TimeInterval     time.Duration
	PrettyLatency    bool
	enableLatency    bool
}

// defaultConfig is the default config
var defaultLoggerMiddlewareConfig = LoggerMiddlewareConfig{
	Next:         nil,
	Format:       []string{TagTime, TagStatus, TagLatency, TagMethod, TagPath},
	TimeFormat:   time.RFC3339,
	TimeZone:     "Local",
	TimeInterval: timeIntervalDefault * time.Millisecond,
}

func (cfg *LoggerMiddlewareConfig) setTimezoneLocation() {
	tz, err := time.LoadLocation(cfg.TimeZone)
	if err != nil || tz == nil {
		cfg.timeZoneLocation = time.Local
	} else {
		cfg.timeZoneLocation = tz
	}
}

func (cfg *LoggerMiddlewareConfig) checkEnableLatency() {
	for _, tag := range cfg.Format {
		if tag == TagLatency {
			cfg.enableLatency = true
			break
		}
	}
}

func setLoggerMiddlewareConfig(config []LoggerMiddlewareConfig) LoggerMiddlewareConfig {
	// Return default config if nothing provided
	if len(config) < 1 {
		cfg := defaultLoggerMiddlewareConfig
		cfg.logger = *logger
		return cfg
	}

	// Override default config
	cfg := config[0]
	if cfg.Next == nil {
		cfg.Next = defaultLoggerMiddlewareConfig.Next
	}
	if cfg.Format == nil {
		cfg.Format = defaultLoggerMiddlewareConfig.Format
	}
	if cfg.TimeZone == "" {
		cfg.TimeZone = defaultLoggerMiddlewareConfig.TimeZone
	}
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = defaultLoggerMiddlewareConfig.TimeFormat
	}
	if int(cfg.TimeInterval) <= 0 {
		cfg.TimeInterval = defaultLoggerMiddlewareConfig.TimeInterval
	}
	cfg.logger = *logger
	return cfg
}

//nolint:gocyclo
func FiberLoggerMiddleware(config ...LoggerMiddlewareConfig) fiber.Handler {
	cfg := setLoggerMiddlewareConfig(config)
	cfg.setTimezoneLocation()
	cfg.checkEnableLatency()

	// Create correct time format
	var timestamp atomic.Value
	timestamp.Store(time.Now().In(cfg.timeZoneLocation).Format(cfg.TimeFormat))

	// Update date/time in a separate go routine
	for _, tag := range cfg.Format {
		if tag == TagTime {
			go func() {
				for {
					time.Sleep(cfg.TimeInterval)
					timestamp.Store(time.Now().In(cfg.timeZoneLocation).Format(cfg.TimeFormat))
				}
			}()
			break
		}
	}

	// Set PID once
	pid := strconv.Itoa(os.Getpid())

	// Set variables
	var (
		once       sync.Once
		errHandler fiber.ErrorHandler
	)

	// Return new handler
	return func(ctx *fiber.Ctx) error {
		// Don't execute the middleware if Next returns true
		if cfg.Next != nil && cfg.Next(ctx) {
			return ctx.Next()
		}

		// Set error handler once
		once.Do(func() {
			// override error handler
			errHandler = ctx.App().Config().ErrorHandler
		})

		var start, stop time.Time

		// Set latency start time
		if cfg.enableLatency {
			start = time.Now()
		}

		// Handle request, store err for logging
		chainErr := ctx.Next()

		// Manually call error handler
		if chainErr != nil {
			if err := errHandler(ctx, chainErr); err != nil {
				_ = ctx.SendStatus(fiber.StatusInternalServerError)
			}
		}

		// Set latency stop time
		if cfg.enableLatency {
			stop = time.Now()
		}

		status := ctx.Response().StatusCode()

		var event *zerolog.Event
		switch {
		case status >= fiber.StatusContinue && status < fiber.StatusBadRequest:
			event = cfg.logger.Info()
		case status >= fiber.StatusBadRequest && status < fiber.StatusInternalServerError:
			event = cfg.logger.Warn()
		case status >= fiber.StatusInternalServerError:
			event = cfg.logger.Error()
		default:
			event = cfg.logger.Debug()
		}

		for _, tag := range cfg.Format {
			switch tag {
			case TagTime:
				event = event.Str(TagTime, timestamp.Load().(string))
			case TagReferer:
				event = event.Str(TagReferer, ctx.Get(fiber.HeaderReferer))
			case TagProtocol:
				event = event.Str(TagProtocol, ctx.Protocol())
			case TagPid:
				event = event.Str(TagPid, pid)
			case TagID:
				event = event.Str(TagID, ctx.Get(fiber.HeaderXRequestID))
			case TagIP:
				event = event.Str(TagIP, ctx.IP())
			case TagIPs:
				event = event.Str(TagIPs, ctx.Get(fiber.HeaderXForwardedFor))
			case TagHost:
				event = event.Str(TagHost, ctx.Hostname())
			case TagPath:
				event = event.Str(TagPath, ctx.Path())
			case TagURL:
				event = event.Str(TagURL, ctx.OriginalURL())
			case TagUA:
				event = event.Str(TagUA, ctx.Get(fiber.HeaderUserAgent))
			case TagLatency:
				if cfg.PrettyLatency {
					event = event.Str(TagLatency, stop.Sub(start).String())
				} else {
					event = event.Dur(TagLatency, stop.Sub(start))
				}
			case TagBody:
				event = event.Bytes(TagBody, ctx.Body())
			case TagBytesReceived:
				event = event.Int(TagBytesReceived, len(ctx.Request().Body()))
			case TagBytesSent:
				event = event.Int(TagBytesSent, len(ctx.Response().Body()))
			case TagRoute:
				event = event.Str(TagRoute, ctx.Route().Path)
			case TagStatus:
				event = event.Int(TagStatus, status)
			case TagResBody:
				event = event.Bytes(TagResBody, ctx.Response().Body())
			case TagQueryStringParams:
				event = event.Str(TagQueryStringParams, ctx.Request().URI().QueryArgs().String())
			case TagMethod:
				event = event.Str(TagMethod, ctx.Method())
			case TagError:
				if chainErr != nil {
					event = event.Err(chainErr)
				}
			default:
				// Check if we have a value tag i.e.: "header:x-key"
				switch {
				case strings.HasPrefix(tag, TagHeader):
					event = event.Str(tag[7:], ctx.Get(tag[7:]))
				case strings.HasPrefix(tag, TagQuery):
					event = event.Str(tag[6:], ctx.Query(tag[6:]))
				case strings.HasPrefix(tag, TagForm):
					event = event.Str(tag[5:], ctx.FormValue(tag[5:]))
				case strings.HasPrefix(tag, TagCookie):
					event = event.Str(tag[7:], ctx.Cookies(tag[7:]))
				case strings.HasPrefix(tag, TagLocals):
					switch v := ctx.Locals(tag[7:]).(type) {
					case []byte:
						event = event.Bytes(tag[7:], v)
					case string:
						event = event.Str(tag[7:], v)
					case nil:
						// NOOP
					default:
						event = event.Str(tag[7:], fmt.Sprintf("%v", v))
					}
				}
			}
		}

		event.Msg(utils.StatusMessage(status))
		// End chain
		return nil
	}
}
