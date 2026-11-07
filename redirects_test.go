package redirects

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/relative", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/final", http.StatusFound)
	})
	mux.HandleFunc("/temporary", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/permanent", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("/permanent", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/final", http.StatusPermanentRedirect)
	})
	mux.HandleFunc("/loop", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/loop", http.StatusFound)
	})
	mux.HandleFunc("/no-location", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusFound)
	})
	mux.HandleFunc("/final", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	s := httptest.NewServer(mux)
	t.Cleanup(s.Close)
	return s
}

func TestGet(t *testing.T) {
	s := newTestServer(t)

	tests := []struct {
		name      string
		path      string
		wantCodes []int
	}{
		{"relative location", "/relative", []int{302, 200}},
		{"307 and 308", "/temporary", []int{307, 308, 200}},
		{"https in query string", "/final?next=https://example.com", []int{200}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Get(s.URL + tt.path)
			if d.Error {
				t.Fatalf("unexpected error: %s", d.ErrorMessage)
			}
			if len(d.Redirects) != len(tt.wantCodes) {
				t.Fatalf("got %d hops, want %d", len(d.Redirects), len(tt.wantCodes))
			}
			for i, hop := range d.Redirects {
				if hop.Number != i {
					t.Errorf("hop %d has number %d", i, hop.Number)
				}
				if hop.StatusCode != tt.wantCodes[i] {
					t.Errorf("hop %d has status %d, want %d", i, hop.StatusCode, tt.wantCodes[i])
				}
			}
			if last := d.Redirects[len(d.Redirects)-1]; !strings.HasPrefix(last.URL, s.URL+"/final") {
				t.Errorf("final URL is %s", last.URL)
			}
		})
	}
}

func TestGetErrors(t *testing.T) {
	s := newTestServer(t)

	tests := []struct {
		name    string
		url     string
		wantErr string
	}{
		{"empty URL", "", "empty URL"},
		{"redirect loop", s.URL + "/loop", "stopped after 20 redirects"},
		{"missing location", s.URL + "/no-location", "Location header"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Get(tt.url)
			if !d.Error || !strings.Contains(d.ErrorMessage, tt.wantErr) {
				t.Errorf("got error %v %q, want %q", d.Error, d.ErrorMessage, tt.wantErr)
			}
		})
	}
}

func TestBlockPrivateIPs(t *testing.T) {
	s := newTestServer(t)

	d := GetWithOptions(s.URL+"/final", Options{BlockPrivateIPs: true})
	if !d.Error || !strings.Contains(d.ErrorMessage, "not a public address") {
		t.Errorf("got error %v %q, want blocked", d.Error, d.ErrorMessage)
	}

	for _, addr := range []string{"127.0.0.1:80", "10.0.0.1:80", "192.168.1.1:443", "169.254.169.254:80", "100.64.0.1:80", "[::1]:80", "[fd00::1]:80", "[::ffff:127.0.0.1]:80"} {
		if blockPrivateIPs("tcp", addr, nil) == nil {
			t.Errorf("%s was not blocked", addr)
		}
	}
	for _, addr := range []string{"1.1.1.1:443", "[2606:4700:4700::1111]:443"} {
		if err := blockPrivateIPs("tcp", addr, nil); err != nil {
			t.Errorf("%s was blocked: %v", addr, err)
		}
	}
}
