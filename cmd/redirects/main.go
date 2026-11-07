package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/sbroekhoven/redirects"
)

// version is set at build time with -ldflags "-X main.version=v1.2.3".
var version = ""

func main() {
	jsonFlag := flag.Bool("json", false, "output as JSON")
	mdFlag := flag.Bool("markdown", false, "output as Markdown table")
	noColorFlag := flag.Bool("no-color", false, "disable colored output")
	versionFlag := flag.Bool("version", false, "print version and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: redirects [options] <url>\n\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	// Allow flags after the URL too: redirects example.com -json
	args := flag.Args()
	if len(args) > 1 {
		flag.CommandLine.Parse(args[1:])
		args = append(args[:1], flag.Args()...)
	}

	if *versionFlag {
		fmt.Println("redirects", getVersion())
		return
	}

	if len(args) != 1 {
		flag.Usage()
		os.Exit(2)
	}

	data := redirects.Get(args[0])

	switch {
	case *jsonFlag:
		printJSON(os.Stdout, data)
	case *mdFlag:
		printMarkdown(os.Stdout, data)
	default:
		p := painter{enabled: useColor(*noColorFlag)}
		printText(os.Stdout, data, p)
	}

	if data.Error {
		os.Exit(1)
	}
}

func getVersion() string {
	if version != "" {
		return version
	}
	// Installed with "go install ...@vX.Y.Z": Go records the module version
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "dev"
}

func printJSON(w io.Writer, data *redirects.Data) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(data)
}

func printText(w io.Writer, data *redirects.Data, p painter) {
	hops := data.Redirects

	// Column widths, measured on the plain text so colors don't break alignment
	numW, statusW, protoW, tlsW := len("#"), len("STATUS"), len("PROTO"), len("TLS")
	for _, r := range hops {
		numW = max(numW, len(fmt.Sprint(r.Number+1)))
		statusW = max(statusW, len(statusText(r.StatusCode)))
		protoW = max(protoW, len(r.Protocol))
		tlsW = max(tlsW, len(tlsText(r.TLSVersion)))
	}

	if len(hops) > 0 {
		fmt.Fprintf(w, "%s  %s  %s  %s  %s\n", p.dim(pad("#", numW)), p.dim(pad("STATUS", statusW)),
			p.dim(pad("PROTO", protoW)), p.dim(pad("TLS", tlsW)), p.dim("URL"))
	}
	for _, r := range hops {
		fmt.Fprintf(w, "%s  %s  %s  %s  %s\n",
			pad(fmt.Sprint(r.Number+1), numW),
			p.status(r.StatusCode, pad(statusText(r.StatusCode), statusW)),
			pad(r.Protocol, protoW),
			p.tls(r.TLSVersion, pad(tlsText(r.TLSVersion), tlsW)),
			r.URL)
		if r.InsecureDowngrade {
			fmt.Fprintf(w, "%s  %s\n", strings.Repeat(" ", numW),
				p.red("⚠ insecure downgrade: HTTPS redirected to HTTP"))
		}
	}

	if data.Error {
		if len(hops) > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(os.Stderr, "%s %s\n", p.red("error:"), data.ErrorMessage)
		return
	}

	if len(hops) == 0 {
		return
	}
	final := hops[len(hops)-1]
	n := len(hops) - 1
	noun := "redirects"
	if n == 1 {
		noun = "redirect"
	}
	fmt.Fprintf(w, "\n%s %s → %s\n", p.dim(fmt.Sprintf("%d %s,", n, noun)),
		p.status(final.StatusCode, statusText(final.StatusCode)), p.bold(final.URL))
}

func printMarkdown(w io.Writer, data *redirects.Data) {
	fmt.Fprintf(w, "# Redirects for %s\n\n", data.URL)
	fmt.Fprintln(w, "| # | Status | Protocol | TLS | URL |")
	fmt.Fprintln(w, "|---|--------|----------|-----|-----|")
	for _, r := range data.Redirects {
		warning := ""
		if r.InsecureDowngrade {
			warning = " ⚠️ insecure downgrade"
		}
		fmt.Fprintf(w, "| %d | %s | %s | %s | `%s`%s |\n",
			r.Number+1, statusText(r.StatusCode), r.Protocol, tlsText(r.TLSVersion), r.URL, warning)
	}
	if data.Error {
		fmt.Fprintf(w, "\n**Error:** %s\n", data.ErrorMessage)
	}
}

func statusText(code int) string {
	return strings.TrimSpace(fmt.Sprintf("%d %s", code, http.StatusText(code)))
}

func tlsText(v string) string {
	if v == "" || v == "N/A" {
		return "none"
	}
	return v
}

func pad(s string, width int) string {
	return s + strings.Repeat(" ", max(0, width-len([]rune(s))))
}

// useColor follows the NO_COLOR convention (https://no-color.org) and only
// colors output that goes to a terminal.
func useColor(disabled bool) bool {
	if disabled || os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	fi, err := os.Stdout.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

type painter struct{ enabled bool }

func (p painter) wrap(code, s string) string {
	if !p.enabled {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func (p painter) bold(s string) string { return p.wrap("1", s) }
func (p painter) dim(s string) string  { return p.wrap("2", s) }
func (p painter) red(s string) string  { return p.wrap("31", s) }

func (p painter) status(code int, s string) string {
	switch {
	case code >= 200 && code < 300:
		return p.wrap("32", s) // green
	case code >= 300 && code < 400:
		return p.wrap("33", s) // yellow
	default:
		return p.wrap("31", s) // red
	}
}

func (p painter) tls(v, s string) string {
	if tlsText(v) == "none" {
		return p.dim(s)
	}
	return s
}
