package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/sbroekhoven/redirects"
)

func main() {
	// Define command-line flags
	urlFlag := flag.String("url", "", "The URL to follow redirects for")

	// Parse command-line flags
	flag.Parse()

	// Validate the URL flag
	if *urlFlag == "" {
		fmt.Println("Usage: redirects -url <URL>")
		os.Exit(1)
	}

	// Call the Get function from the redirects package
	data := redirects.Get(*urlFlag)

	// Check for errors
	if data.Error {
		log.Fatalf("Error: %s\n", data.ErrorMessage)
	}

	// Print the results
	fmt.Printf("URL: %s\n", data.URL)
	for _, redirect := range data.Redirects {
		fmt.Printf("Redirect %d: %s (Status Code: %d, Protocol: %s, TLS Version: %s)\n",
			redirect.Number, redirect.URL, redirect.StatusCode, redirect.Protocol, redirect.TLSVersion)
	}
}
