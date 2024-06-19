package service

import (
	"context"

	"errors"
	"net/http"
	"net/http/pprof"

	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

func startProfiler(ctx context.Context, addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		err := server.Shutdown(context.Background())
		if err != nil {
			panic("pprof server shutdown failed: " + err.Error())
		}
	}()

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic("pprof server failed: " + err.Error())
		}
	}()
}

func startDataDogProfiler(ctx context.Context) {
	opts := make([]profiler.Option, 0)

	opts = append(opts, profiler.WithProfileTypes(
		profiler.CPUProfile,
		profiler.HeapProfile,
		// higher overhead
		profiler.BlockProfile,
		profiler.MutexProfile,
		profiler.GoroutineProfile,
	))

	err := profiler.Start(opts...)
	if err != nil {
		panic("failed to start DataDog profiler: " + err.Error())
	}

	go func() {
		<-ctx.Done()
		profiler.Stop()
	}()
}
