package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sbroekhoven/redirects"
)

func main() {
	jsonFlag := flag.Bool("json", false, "output as JSON")
	mdFlag := flag.Bool("markdown", false, "output as Markdown table")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: redirects [options] <url>\n\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	url := flag.Arg(0)
	data := redirects.Get(url)

	if data.Error {
		fmt.Fprintf(os.Stderr, "Error: %s\n", data.ErrorMessage)
		os.Exit(1)
	}

	switch {
	case *jsonFlag:
		printJSON(data)
	case *mdFlag:
		printMarkdown(data)
	default:
		printText(data)
	}
}

func printJSON(data *redirects.Data) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(data)
}

func printText(data *redirects.Data) {
	fmt.Printf("URL: %s\n", data.URL)
	fmt.Printf("%-4s  %-6s  %-8s  %-10s  %s\n", "#", "Code", "Proto", "TLS", "URL")
	fmt.Println(strings.Repeat("-", 80))
	for _, r := range data.Redirects {
		warning := ""
		if r.InsecureDowngrade {
			warning = " [INSECURE DOWNGRADE: https -> http]"
		}
		fmt.Printf("%-4d  %-6d  %-8s  %-10s  %s%s\n",
			r.Number, r.StatusCode, r.Protocol, r.TLSVersion, r.URL, warning)
	}
}

func printMarkdown(data *redirects.Data) {
	fmt.Printf("# Redirects for %s\n\n", data.URL)
	fmt.Println("| # | Code | Proto | TLS | URL |")
	fmt.Println("|---|------|-------|-----|-----|")
	for _, r := range data.Redirects {
		warning := ""
		if r.InsecureDowngrade {
			warning = " ⚠️ insecure downgrade"
		}
		fmt.Printf("| %d | %d | %s | %s | %s%s |\n",
			r.Number, r.StatusCode, r.Protocol, r.TLSVersion, r.URL, warning)
	}
}
