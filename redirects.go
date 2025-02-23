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

// Get function
// This function takes a URL and a nameserver as arguments and returns a struct
// with information about the URL and the redirects it goes through.
// The function will follow a maximum of 20 redirects.
// If there is an error when making the request, the function will return an error
// message.
// If a redirect goes to an invalid URL, the function will not return an error,
// but instead will set the Error field of the returned struct to true and the
// ErrorMessage field to a string with the error message.
// The TLS version is currently not really relevant, but it is included in the
// returned struct.
func Get(redirecturl string, nameserver string) *Data {

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

	// Loop through up to 20 redirects
	for i := 0; i < maxRedirects; i++ {

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
		req.Header.Set("User-Agent", "Mozilla/5.0 (Golang_Research_Bot/3.0)")

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
		// if redirect == nil {
		// 	// If the Redirects struct is nil, set the Error field to true and the
		// 	// ErrorMessage field to the error message.
		// 	r.Error = true
		// 	r.ErrorMessage = "redirect == nil"
		// 	return r
		// }

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
