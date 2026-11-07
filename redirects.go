package redirects

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"syscall"
	"time"
)

// Data struct
type Data struct {
	URL          string       `json:"url,omitempty"`
	Redirects    []*Redirects `json:"redirects,omitempty"`
	Error        bool         `json:"error,omitempty"`
	ErrorMessage string       `json:"errormessage,omitempty"`
}

// Redirects struct
type Redirects struct {
	Number            int    `json:"number"`
	StatusCode        int    `json:"statuscode,omitempty"`
	URL               string `json:"url,omitempty"`
	Protocol          string `json:"protocol,omitempty"`
	TLSVersion        string `json:"tlsversion,omitempty"` // Dont know if TLS version stays.
	InsecureDowngrade bool   `json:"insecure_downgrade,omitempty"`
}

const maxRedirects = 20

// Options changes how GetWithOptions follows redirects.
type Options struct {
	// BlockPrivateIPs refuses to connect to loopback, private, link-local and
	// other non-public addresses. Use this when the URL comes from untrusted
	// users, for example in a web service.
	BlockPrivateIPs bool
}

// Get follows the redirects of a URL and returns every hop.
func Get(redirecturl string) *Data {
	return GetWithOptions(redirecturl, Options{})
}

// GetWithOptions follows the redirects of a URL and returns every hop.
func GetWithOptions(redirecturl string, opts Options) *Data {

	r := new(Data)

	r.URL = redirecturl

	err := validateURL(redirecturl)
	if err != nil {
		r.Error = true
		r.ErrorMessage = err.Error()
		return r
	}

	// Add http:// to url if missing
	if !hasHTTPScheme(redirecturl) {
		// TODO: Set warning
		redirecturl = "http://" + redirecturl
	}

	// Create a new HTTP client
	client := createHTTPClient(opts)

	previousURL := ""

	// Loop through up to 20 redirects
	for i := range maxRedirects {

		// Prepare the request
		req, err := http.NewRequest("GET", redirecturl, nil)
		if err != nil {
			// If there is an error with the request, set the Error field to true
			// and the ErrorMessage field to the error message.
			r.Error = true
			r.ErrorMessage = err.Error()
			return r
		}

		// Set the User-Agent header
		req.Header.Set("User-Agent", "Mozilla/5.0 (https://github.com/sbroekhoven/redirects)")

		// Do the request
		resp, err := client.Do(req)
		if err != nil {
			// If there is an error with the request, set the Error field to true
			// and the ErrorMessage field to the error message.
			r.Error = true
			r.ErrorMessage = err.Error()
			return r
		}
		// Only the headers are needed, so close the body right away
		resp.Body.Close()

		// Create a new Redirects struct
		redirect := new(Redirects)

		// Set the fields of the Redirects struct
		redirect.Number = i
		redirect.StatusCode = resp.StatusCode
		redirect.URL = resp.Request.URL.String()
		redirect.Protocol = resp.Proto

		// If the connection used TLS, get the TLS version
		if resp.TLS != nil {
			redirect.TLSVersion = tls.VersionName(resp.TLS.Version)
		} else {
			redirect.TLSVersion = "N/A"
		}

		// Flag if a previous https hop redirected down to http
		if previousURL != "" && strings.HasPrefix(previousURL, "https://") && strings.HasPrefix(redirecturl, "http://") {
			redirect.InsecureDowngrade = true
		}

		previousURL = redirecturl

		// Add the Redirects struct to the slice of Redirects structs
		r.Redirects = append(r.Redirects, redirect)

		// Any status other than a redirect is the final destination
		if !isRedirect(resp.StatusCode) {
			return r
		}

		// Resolve the Location header against the current URL, so relative
		// redirects like "/path" work too
		location, err := resp.Location()
		if err != nil {
			r.Error = true
			r.ErrorMessage = "invalid or missing Location header: " + err.Error()
			return r
		}
		redirecturl = location.String()
	}

	r.Error = true
	r.ErrorMessage = fmt.Sprintf("stopped after %d redirects", maxRedirects)
	return r
}

func isRedirect(statusCode int) bool {
	switch statusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return true
	}
	return false
}

func hasHTTPScheme(u string) bool {
	u = strings.ToLower(u)
	return strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://")
}

func createHTTPClient(opts Options) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if opts.BlockPrivateIPs {
		// Check the address after DNS resolution, so hostnames that resolve
		// to internal addresses are blocked too
		dialer := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			Control:   blockPrivateIPs,
		}
		transport.DialContext = dialer.DialContext
		// A proxy would hide the real destination from the check above
		transport.Proxy = nil
	}

	return &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 30 * time.Second,
	}
}

// carrierGradeNAT is 100.64.0.0/10, which netip doesn't treat as private
var carrierGradeNAT = netip.MustParsePrefix("100.64.0.0/10")

func blockPrivateIPs(network, address string, _ syscall.RawConn) error {
	addrPort, err := netip.ParseAddrPort(address)
	if err != nil {
		return err
	}
	ip := addrPort.Addr().Unmap()
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || carrierGradeNAT.Contains(ip) {
		return fmt.Errorf("connection to %s blocked: not a public address", ip)
	}
	return nil
}

func validateURL(redirecturl string) error {

	if redirecturl == "" {
		return errors.New("empty URL")
	}
	// Parse the URL using the url.Parse() function
	_, err := url.Parse(redirecturl)
	if err != nil {
		// If the URL is invalid, return the error
		return err
	}
	// If the URL is valid, return nil
	return nil
}
