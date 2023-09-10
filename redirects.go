package redirects

import (
	"net/http"
	"net/url"
	"strings"
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
	Number     int    `json:"number"`
	StatusCode int    `json:"statuscode,omitempty"`
	URL        string `json:"url,omitempty"`
	Protocol   string `json:"protocol,omitempty"`
	TLSVersion uint16 `json:"tlsversion,omitempty"`
}

// Get function
func Get(redirecturl string, nameserver string) *Data {
	r := new(Data)

	r.URL = redirecturl
	_, err := url.Parse(redirecturl)
	if err != nil {
		r.Error = true
		r.ErrorMessage = err.Error()
		return r
	}

	var i int

	// we want to follow max 20 redirects
	for i < 20 {
		// set client to CheckRedirect, not following the redirect
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}}

		// redirecturl prefix check for incomplete
		if !caseInsenstiveContains(redirecturl, "http://") && !caseInsenstiveContains(redirecturl, "https://") {
			// TODO: Set warning
			redirecturl = "http://" + redirecturl
		}

		// Prepare the request
		req, err := http.NewRequest("GET", redirecturl, nil)
		if err != nil {
			r.Error = true
			r.ErrorMessage = err.Error()
			return r
		}

		// Set User-Agent
		req.Header.Set("User-Agent", "Golang_Research_Bot/3.0")

		// Do the request.
		resp, err := client.Do(req)
		if err != nil {
			r.Error = true
			r.ErrorMessage = err.Error()
			return r
		}
		defer resp.Body.Close()

		// Set soms vars.
		redirect := new(Redirects)
		redirect.Number = i
		redirect.StatusCode = resp.StatusCode
		redirect.URL = resp.Request.URL.String()
		redirect.Protocol = resp.Proto
		redirect.TLSVersion = resp.TLS.Version

		r.Redirects = append(r.Redirects, redirect)

		if resp.StatusCode == 200 || resp.StatusCode > 303 {
			break
		} else {
			redirecturl = resp.Header.Get("Location")
			i++
		}
	}

	return r
}

func caseInsenstiveContains(a, b string) bool {
	return strings.Contains(strings.ToUpper(a), strings.ToUpper(b))
}
