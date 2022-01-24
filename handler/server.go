package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	headersSep = ", "
)

var allowedHeaders = []string{
	"Accept",
	"Content-Type",
	"Content-Length",
	"Cookie",
	"Accept-Encoding",
	"Authorization",
	"X-CSRF-Token",
	"X-Requested-With",
	"X-Forwarded-For",
	"CF-Connecting-IP",
	"CF-Real-IP",
}

var allowedMethods = []string{
	"OPTIONS",
	"GET",
	"PUT",
	"PATCH",
	"POST",
	"DELETE",
}

type Handler struct {
	logger *zap.SugaredLogger
	tracer opentracing.Tracer
	closer io.Closer
	router *mux.Router
}

func (h *Handler) initLogger() error {
	cfg := zap.NewDevelopmentConfig()
	cfg.Development = viper.GetString("env") != "main"
	cfg.DisableStacktrace = !viper.GetBool("log_stacktrace")

	if viper.GetBool("log_json") {
		cfg.Encoding = "json"
	} else {
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

	l, err := cfg.Build()
	if err != nil {
		return err
	}

	h.logger = l.Sugar()

	return nil
}

func (h *Handler) initTracer() error {
	jaegerUrl := viper.GetString("jaeger_trace")
	if len(jaegerUrl) > 0 {
		jaegerSrvName := fmt.Sprintf("busybox-%s", viper.GetString("env"))
		address := viper.GetString("jaeger_addr")
		traceTransport, err := jaeger.NewUDPTransport(address, 0)
		if err != nil {
			h.logger.Errorf("Unable to init tracing transport: %s", err)
			return err
		}

		tracer, closer, err := config.Configuration{
			ServiceName: jaegerSrvName,
		}.NewTracer(
			config.Sampler(jaeger.NewConstSampler(true)),
			config.Reporter(jaeger.NewRemoteReporter(traceTransport, jaeger.ReporterOptions.Logger(jaeger.StdLogger))),
			config.Metrics(prometheus.New()),
		)
		if err != nil {
			h.logger.Errorf("Unable to start tracer: %s", err)
			return err
		}

		h.logger.Debugw("Jaeger tracing client initialized", "collector_url", address)
		h.tracer = tracer
		h.closer = closer
	}

	return nil
}

func (h *Handler) Run() error {
	if err := h.initLogger(); err != nil {
		return err
	}

	if err := h.initTracer(); err != nil {
		return err
	}

	r := mux.NewRouter()

	r.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)
	r.HandleFunc("/health", h.healthCheck).Methods(http.MethodGet)
	r.HandleFunc("/debug", h.mainHandler).Methods(http.MethodGet)

	h.router = r

	listenAddr := viper.GetString("listen_addr")
	h.logger.Infow("Starting HTTP Server", "listen_addr", listenAddr)
	return http.ListenAndServe(listenAddr, h)
}

func (h *Handler) GracefulShutdown(sig string) {
	if h.logger != nil {
		h.logger.Warnf("Shutdown signal '%s' received", sig)
	}

	if h.closer != nil {
		h.closer.Close()
	}

	os.Exit(0)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Set request headers for AJAX requests
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, headersSep))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, headersSep))
	}

	// handle preflight request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx = context.WithValue(ctx, "host", r.Host)
	ctx = context.WithValue(ctx, "path", r.URL.Path)
	ctx = context.WithValue(ctx, "remote_addr", r.RemoteAddr)
	ctx = context.WithValue(ctx, "x_forwarded_for", r.Header.Get("X-Forwarded-For"))

	if h.tracer != nil {
		span := h.tracer.StartSpan(strings.TrimPrefix(r.URL.Path, "/"))
		span.SetTag("host", r.Host)
		span.SetTag("method", r.Method)
		span.SetTag("path", r.URL.Path)
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	h.logger.Infow("Handling request", "method", r.Method, "path", r.URL.Path, "query", r.URL.RawQuery)

	h.router.ServeHTTP(w, r.WithContext(ctx))
}

func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, map[string]interface{}{
		"healthy":   true,
		"timestamp": time.Now().Format(time.RFC1123),
	})
}

func (h *Handler) mainHandler(w http.ResponseWriter, r *http.Request) {
	var result []interface{}
	for name, values := range r.Header {
		h.logger.Infof("%s: %s", name, values)
		result = append(result, map[string]interface{}{
			"name":  name,
			"value": r.Header.Values(name),
		})
	}

	writeResponse(w, result)
}

func writeResponse(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	response, err := json.Marshal(v)
	if err != nil {
		w.Write([]byte("Cannot marshal response: " + err.Error()))
		return
	}

	w.WriteHeader(200)
	w.Write(response)
}
