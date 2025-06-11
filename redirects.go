package redirects

import (
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"
	"strings"
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
	Number     int    `json:"number"`
	StatusCode int    `json:"statuscode,omitempty"`
	URL        string `json:"url,omitempty"`
	Protocol   string `json:"protocol,omitempty"`
	TLSVersion string `json:"tlsversion,omitempty"` // Dont know if TLS version stays.
}

const maxRedirects = 20

func Get(redirecturl string) *Data {

	r := new(Data)

	r.URL = redirecturl

	err := validateURL(redirecturl)
	if err != nil {
		r.Error = true
		r.ErrorMessage = err.Error()
		return r
	}

	// Create a new HTTP client
	client := createHTTPClient()

	// Create a slice of integers from 0 to maxRedirects-1
	redirectIndices := make([]int, maxRedirects)
	for i := range redirectIndices {
		redirectIndices[i] = i
	}

	// Loop through up to 20 redirects using range
	for i := range redirectIndices {

		// Set the client to follow redirects, but not to follow the redirect
		// automatically. Instead, the redirect will be stored in the Location
		// header and the client will stop following redirects.
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}

		// Add http:// to url if missing
		if !caseInsenstiveContains(redirecturl, "http://") && !caseInsenstiveContains(redirecturl, "https://") {
			// TODO: Set warning
			redirecturl = "http://" + redirecturl
		}

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
		defer resp.Body.Close()

		// Create a new Redirects struct
		redirect := new(Redirects)

		// Set the fields of the Redirects struct
		redirect.Number = i
		redirect.StatusCode = resp.StatusCode
		redirect.URL = resp.Request.URL.String()
		redirect.Protocol = resp.Proto

		// If the URL is an https URL, get the TLS version
		if caseInsenstiveContains(redirecturl, "https://") {
			redirect.TLSVersion = tls.VersionName(resp.TLS.Version)
		} else {
			redirect.TLSVersion = "N/A"
		}

		// Add the Redirects struct to the slice of Redirects structs
		r.Redirects = append(r.Redirects, redirect)

		// If the status code is 200 or greater than 303, break out of the loop
		if resp.StatusCode == 200 || resp.StatusCode > 303 {
			break
		} else {
			if len(resp.Header.Get("Location")) > 0 {
				redirecturl = resp.Header.Get("Location")
			} else if len(resp.Header.Get("location")) > 0 {
				redirecturl = resp.Header.Get("location")
			} else if len(resp.Header.Get("LOCATION")) > 0 {
				redirecturl = resp.Header.Get("LOCATION")
			} else {
				r.Error = true
				r.ErrorMessage = "Location header is empty"
				return r
			}

			// Ensure redirecturl is fully qualified
			if !strings.HasPrefix(redirecturl, "http://") && !strings.HasPrefix(redirecturl, "https://") {
				redirecturl = "http://" + redirecturl
			}
		}
	}

	// Return the Data struct
	return r
}

func caseInsenstiveContains(a, b string) bool {
	return strings.Contains(strings.ToUpper(a), strings.ToUpper(b))
}

func createHTTPClient() *http.Client {
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 30 * time.Second,
	}
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
