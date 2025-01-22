package logging

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"go.elastic.co/apm/module/apmfasthttp/v2"
	"go.elastic.co/apm/module/apmhttp/v2"
	"go.elastic.co/apm/v2"
	"jira-clone-api/utilities/local"
)

const txKey = "apmfasthttp_transaction"

// FiberApmMiddleware returns a new Fiber middleware handler for tracing
// requests and reporting errors.
//
// This middleware will recover and report panics, so it can
// be used instead of default recover middleware.
//
// By default, the middleware will use apm.DefaultTracer.
// Use WithTracer to specify an alternative tracer.
// Use WithPanicPropagation to disable panic recover.
func FiberApmMiddleware(o ...Option) fiber.Handler {
	m := &fiberApmMiddleware{
		tracer:           apm.DefaultTracer(),
		requestIgnorer:   apmfasthttp.NewDynamicServerRequestIgnorer(apm.DefaultTracer()),
		panicPropagation: false,
	}

	for _, o := range o {
		o(m)
	}
	return m.handle
}

type fiberApmMiddleware struct {
	tracer           *apm.Tracer
	requestIgnorer   apmfasthttp.RequestIgnorerFunc
	panicPropagation bool
}

type txCloser struct {
	tx *apm.Transaction
	bc *apm.BodyCapturer
}

func init() {
	origTransactionFromContext := apm.OverrideTransactionFromContext
	apm.OverrideTransactionFromContext = func(ctx context.Context) *apm.Transaction {
		if tx, ok := ctx.Value(txKey).(*txCloser); ok {
			return tx.tx
		}
		return origTransactionFromContext(ctx)
	}

	origBodyCapturerFromContext := apm.OverrideBodyCapturerFromContext
	apm.OverrideBodyCapturerFromContext = func(ctx context.Context) *apm.BodyCapturer {
		if tx, ok := ctx.Value(txKey).(*txCloser); ok {
			return tx.bc
		}
		return origBodyCapturerFromContext(ctx)
	}
}

func (m *fiberApmMiddleware) handle(c *fiber.Ctx) error {
	reqCtx := c.Context()
	if !m.tracer.Recording() || m.requestIgnorer(reqCtx) {
		return c.Next()
	}

	name := string(reqCtx.Method()) + " " + c.Path()
	tx, body, err := apmfasthttp.StartTransactionWithBody(reqCtx, m.tracer, name)
	if err != nil {
		reqCtx.Error(err.Error(), fiber.StatusInternalServerError)
		return err
	}
	reqCtx.SetUserValue(txKey, newTxCloser(tx, body))
	defer func() {
		resp := c.Response()
		statusCode := local.New(c).GetStatusCode()
		path := c.Route().Path
		if path == "/" && statusCode == fiber.StatusNotFound {
			tx.Name = string(reqCtx.Method()) + " unknown route"
		} else {
			// Workaround for set tx.Name as template path, not absolute
			tx.Name = string(reqCtx.Method()) + " " + path
		}

		if v := recover(); v != nil {
			if m.panicPropagation {
				defer panic(v)
			}

			e := m.tracer.Recovered(v)
			e.SetTransaction(tx)
			setContext(&e.Context, resp, c)
			e.Send()

			c.Status(fiber.StatusInternalServerError)
		}
		tx.Result = apmhttp.StatusCodeResult(statusCode)
		if tx.Sampled() {
			setContext(&tx.Context, resp, c)
		}

		body.Discard()
	}()

	nextErr := c.Next()
	if nextErr != nil {
		resp := c.Response()
		e := m.tracer.NewError(nextErr)
		e.Handled = true
		e.SetTransaction(tx)
		setContext(&e.Context, resp, c)
		e.Send()
	}
	return nextErr
}

func setContext(ctx *apm.Context, resp *fiber.Response, c *fiber.Ctx) {
	localService := local.New(c)
	ctx.SetFramework("fiber", fiber.Version)

	headers := make(http.Header)
	resp.Header.VisitAll(func(k, v []byte) {
		sk := string(k)
		sv := string(v)
		headers.Set(sk, sv)
	})
	ctx.SetHTTPResponseHeaders(headers)
	ctx.SetHTTPStatusCode(localService.GetStatusCode())
	ctx.SetCustom("username", localService.GetUser().Username)
	ctx.SetCustom("extra_body", localService.GetExtraBody())
	ctx.SetLabel("username", localService.GetUser().Username)
}

// newTxCloser returns a transaction closer.
func newTxCloser(tx *apm.Transaction, bc *apm.BodyCapturer) *txCloser {
	return &txCloser{
		tx: tx,
		bc: bc,
	}
}

// Close ends the transaction. The closer exists because, while the wrapped
// handler may return, an underlying streaming body writer may still exist.
// References to the transaction and bodycloser are held in a txCloser struct,
// which is added to the context on the request. From the fasthttp
// documentation:
// All the values are removed from ctx after returning from the top
// RequestHandler. Additionally, Close method is called on each value
// implementing io.Closer before removing the value from ctx.
func (c *txCloser) Close() error {
	c.tx.End()
	c.bc.Discard()
	return nil
}

// Option sets options for tracing.
type Option func(*fiberApmMiddleware)

// WithPanicPropagation returns an Option which enable panic propagation.
// Any panic will be recovered and recorded as an error in a transaction, then
// panic will be caused again.
func WithPanicPropagation() Option {
	return func(o *fiberApmMiddleware) {
		o.panicPropagation = true
	}
}

// WithTracer returns an Option which sets t as the tracer
// to use for tracing server requests. If t is nil, using default tracer.
func WithTracer(t *apm.Tracer) Option {
	if t == nil {
		return noopOption
	}

	return func(o *fiberApmMiddleware) {
		o.tracer = t
	}
}

// WithRequestIgnorer returns an Option which sets fn as the
// function to use to determine whether or not a request should
// be ignored. If fn is nil, using default RequestIgnorerFunc.
func WithRequestIgnorer(fn apmfasthttp.RequestIgnorerFunc) Option {
	if fn == nil {
		return noopOption
	}

	return func(o *fiberApmMiddleware) {
		o.requestIgnorer = fn
	}
}

func noopOption(_ *fiberApmMiddleware) {}
