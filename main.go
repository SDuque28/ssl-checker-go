package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)
// Structs to parse Info JSON responses from SSL Labs API
type Info struct {
	Version      		 string `json:"version"`
	CriteriaVersion 	 string `json:"criteriaVersion"`
	MaxAssessments 		 int    `json:"maxAssessments"`
	CurrentAssessments 	 int    `json:"currentAssessments"`
	NewAssessmentCoolOff int64  `json:"newAssessmentCoolOff"`
	Messages   		   []string `json:"messages"`
}
// Structs to parse Host JSON responses from SSL Labs API
type Host struct {
	Host 	  		string 	 `json:"host"`
	Port      		int    	 `json:"port"`
	Protocol  		string 	 `json:"protocol"`
	IsPublic  		bool   	 `json:"isPublic"`
	Status    		string 	 `json:"status"`
	StatusMessage 	string 	 `json:"statusMessage"`
	StartTime  		int64  	 `json:"startTime"`
	TestTime   		int64  	 `json:"testTime"`
	EngineVersion 	string 	 `json:"engineVersion"`
	CriteriaVersion string 	 `json:"criteriaVersion"`
	Endpoints     []Endpoint `json:"endpoints"`
}
// Structs to parse Endpoint JSON responses from SSL Labs API
type Endpoint struct {
	IpAddress         string `json:"ipAddress"`
	ServerName        string `json:"serverName"`
	StatusMessage     string `json:"statusMessage"`
	StatusDetails  	  string `json:"statusDetails"`
	Grade             string `json:"grade"`
	GradeTrustIgnored string `json:"gradeTrustIgnored"`
	HasWarnings       bool   `json:"hasWarnings"`
	Progress          int    `json:"progress"`
	Duration          int    `json:"duration"`
	Eta 		 	  int    `json:"eta"`
}
// SSLClient struct to interact with SSL Labs API as a client
type SSLClient struct {
	baseurl   string
	client	*http.Client
}
// NewSSLClient initializes and returns a new SSLClient
func NewSSLClient() *SSLClient {
	return &SSLClient{
		baseurl: "https://api.ssllabs.com/api/v2",
		client: &http.Client{Timeout: 30*time.Second},
	}
}
// CheckApiStatus checks the status of the SSL Labs API
func (s *SSLClient) CheckApiStatus() (*Info, error) {
	// Make a GET request to the /info endpoint
	resp, err := s.client.Get(s.baseurl + "/info")
	if err != nil {
		return nil, fmt.Errorf("failed to reach SSL Labs API: %v", err)
	}
	
	defer resp.Body.Close()
	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-OK status: %s", resp.Status)
	}
	// Read and parse the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response: %v", err)
	}
	// Unmarshal JSON into Info struct
	var info Info
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %v", err)
	}
	// Display API status information
	fmt.Println("SSL Labs API is reachable.")
	fmt.Printf("Criteria Version: %s\n", info.CriteriaVersion)
	fmt.Printf("Concurrent assessments allowed: %d\n", info.MaxAssessments)
	fmt.Printf("Current assessments: %d\n", info.CurrentAssessments)
	return &info, nil
}
// StartAssessment initiates a new SSL/TLS assessment for the given domain
func (s *SSLClient) StartAssessment(domain string, publish bool) (*Host, error) {
	url := fmt.Sprintf("%s/analyze?host=%s&all=done&startNew=on", s.baseurl, domain)
	// Append publish parameter if needed
	if publish {
		url += "&publish=on"
	}
	// Make a GET request to start the assessment
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to start assessment: %v", err)
	}
	defer resp.Body.Close()
	// Read and parse the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read assessment response: %v", err)
	}
	// Unmarshal JSON into Host struct
	var host Host
	if err := json.Unmarshal(body, &host); err != nil {
		return nil, fmt.Errorf("failed to parse assessment response: %v", err)
	}
	return &host, nil
}
// CheckAssessmentStatus checks the status of an ongoing assessment for the given domain
func (s *SSLClient) CheckAssessmentStatus(domain string) (*Host, error) {
	url := fmt.Sprintf("%s/analyze?host=%s&all=done", s.baseurl, domain)
	// Make a GET request to check the assessment status
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to check assessment status: %v", err)
	}
	defer resp.Body.Close()
	// Read and parse the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read assessment status response: %v", err)
	}
	// Unmarshal JSON into Host struct
	var host Host
	if err := json.Unmarshal(body, &host); err != nil {
		return nil, fmt.Errorf("failed to parse assessment status response: %v", err)
	}
	return &host, nil
}
// WaitForAssessment polls the assessment status until it is complete
func (s *SSLClient) WaitForAssessment(domain string) (*Host, error) {
	fmt.Println("Waiting for assessment to complete...")
	i, endpoint := 0, 1
	flag := false
	// Poll every 10 seconds until the assessment is complete
	for {
		// Check the current assessment status
		host, err := s.CheckAssessmentStatus(domain)
		if err != nil {
			return nil, fmt.Errorf("failed to check assessment status: %v", err)
		}
		// Display progress for each endpoint
		if !flag || host.Endpoints[i].Progress == 100 {
			if host.Endpoints[i].Progress == 100 {
				fmt.Printf("      %s:%d - %d%%\n",host.Endpoints[i].IpAddress, host.Port, host.Endpoints[i].Progress)
				if i+1 < len(host.Endpoints) {
					i++
				}				
			}
			if endpoint < len(host.Endpoints) + 1{
				fmt.Printf("\n----- PROGRESS ON ENDPOINT %d ----- \n", endpoint)
				endpoint++
			}
			flag = true
		}
		fmt.Printf("      %s:%d - %d%%\n",host.Endpoints[i].IpAddress, host.Port, host.Endpoints[i].Progress)
		// If the status is READY or ERROR, return the host
		if host.Status == "READY" || host.Status == "ERROR" {
			fmt.Println()
			return host, nil
		}
		time.Sleep(10 * time.Second)
	}
}
// displayResults prints the assessment results to the console
func displayResults(host *Host) {
	fmt.Printf("Assessment Results:\n")
	fmt.Printf("Domain: %s\n", host.Host)
	fmt.Printf("Status: %s\n", host.Status)
	// Handle different assessment statuses
	switch host.Status {
		// Display results if the assessment is ready
		case "READY":
			fmt.Printf("Test completed: %s\n", time.Unix(host.TestTime/1000, 0).Format("2006-01-02 15:04:05"))
			// Iterate through each endpoint and display its results
			for i,endpoint := range host.Endpoints {
				fmt.Printf("Endpoint %d:\n", i+1)
				fmt.Printf("  IP Address: %s\n", endpoint.IpAddress)
				fmt.Printf("  Grade: %s\n", endpoint.Grade)
				fmt.Printf("  Status Message: %s\n", endpoint.StatusMessage)
				fmt.Printf("  Has Warnings: %t\n", endpoint.HasWarnings)
				fmt.Println()
			}
		// Display error message if the assessment failed
		case "ERROR":
			fmt.Printf("Assessment failed: %s\n", host.StatusMessage)
	}
}
// main function to parse command-line arguments and run the assessment
func main() {
	// Define command-line flags
	domain := flag.String("domain", "", "Domain to check (e.g., example.com)")
	publish := flag.Bool("publish", false, "Publish results on SSL Labs board")
	help := flag.Bool("help", false, "Show help")
	flag.Parse()
	// Show help if requested or if domain is not provided
	if *help || *domain == "" {
		fmt.Println("SSL Labs API Checker")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(0)
	}
	// Initialize SSLClient
	sslClient := NewSSLClient()
	// Check API status
	info, err := sslClient.CheckApiStatus()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	// Check if maximum concurrent assessments is reached
	if info.CurrentAssessments >= info.MaxAssessments {
		fmt.Println("Maximum number of concurrent assessments reached. Please try again later.")
		os.Exit(1)
	}
	fmt.Printf("Checking SSL/TLS for domain: %s\n", *domain)
	// Start a new assessment
	fmt.Println("Starting Assessment ....")
	host, err := sslClient.StartAssessment(*domain, *publish)
	if err != nil {
		fmt.Printf("Error starting assessment: %v\n", err)
		os.Exit(1)
	}
	// Display initial assessment status
	fmt.Printf("Assessment started for %s\n", host.Host)
	if host.Status != "READY" && host.Status != "ERROR" {
		// Wait for the assessment to complete
		host, err = sslClient.WaitForAssessment(*domain)
		if err != nil {
			fmt.Printf("Error waiting for assessment: %v\n", err)
			os.Exit(1)
		}
	}
	// Display the final results
	displayResults(host)
}
