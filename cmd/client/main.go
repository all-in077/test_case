package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

var serverURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	serverURL = os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	switch os.Args[1] {
	case "list":
		listServers()
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Usage: dns-client add <ip>")
			return
		}
		addServer(os.Args[2])
	case "remove":
		if len(os.Args) < 3 {
			fmt.Println("Usage: dns-client remove <ip>")
			return
		}
		removeServer(os.Args[2])
	case "--help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printHelp()
	}
}

func listServers() {
	resp, err := http.Get(serverURL + "/dns")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Servers []string `json:"servers"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Error parsing response:", err)
		return
	}

	if len(result.Servers) == 0 {
		fmt.Println("No DNS servers found")
		return
	}

	fmt.Println("DNS servers:")
	for _, s := range result.Servers {
		fmt.Println(" -", s)
	}
}

func addServer(ip string) {
	body, _ := json.Marshal(map[string]string{"ip": ip})

	resp, err := http.Post(serverURL+"/dns", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("DNS server %s added successfully\n", ip)
		return
	}

	var result struct {
		Error string `json:"error"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Println("Error:", result.Error)
}

func removeServer(ip string) {
	req, err := http.NewRequest(http.MethodDelete, serverURL+"/dns/"+ip, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("DNS server %s removed successfully\n", ip)
		return
	}

	var result struct {
		Error string `json:"error"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Println("Error:", result.Error)
}

func printHelp() {
	fmt.Println(`DNS Manager CLI

Usage:
  dns-client <command> [arguments]

Commands:
  list            List all DNS servers
  add <ip>        Add a DNS server
  remove <ip>     Remove a DNS server
  --help          Show this help message

Examples:
  dns-client list
  dns-client add 8.8.8.8
  dns-client remove 8.8.8.8`)
}
