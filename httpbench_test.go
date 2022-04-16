package httpbench

import (
	"context"
	"net"
	"net/http"
	"testing"
)

type contextKey struct{}

var value = make(map[string]string, 128)

//go:noinline
func GetValue(ctx context.Context) map[string]string {
	return ctx.Value(contextKey{}).(map[string]string)
}

func handler(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	// ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	// defer cancel()

	ctx = context.WithValue(ctx, contextKey{}, value)

	req = req.WithContext(ctx)
	v := GetValue(req.Context())
	if _, ok := v["non-existant"]; ok {
		rw.Header().Add("X-Huh", "wild")
	}

	rw.WriteHeader(200)
}

func BenchmarkHttpServer(b *testing.B) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer l.Close()

	s := &http.Server{Handler: http.HandlerFunc(handler)}
	defer s.Close()

	addr := "http://" + l.Addr().String() + "/hithere"

	go func() {
		s.Serve(l)
	}()

	baseReq, err := http.NewRequest(http.MethodGet, addr, nil)
	if err != nil {
		b.Fatalf("http.NewRequest: %s", err)
	}

	b.ResetTimer()

	b.ReportAllocs()
	b.SetParallelism(16)
	b.RunParallel(func(pb *testing.PB) {
		c := &http.Client{
			Transport: &http.Transport{},
		}

		var req http.Request
		for pb.Next() {
			req = *baseReq
			resp, err := c.Do(&req)
			if err != nil {
				b.Fatalf("Do: %s", err)
			}
			if resp.StatusCode != http.StatusOK {
				b.Fatalf("non-200 response: %d", resp.StatusCode)
			}
		}
	})
}
