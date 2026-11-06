package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/emil/quickcast/internal/ipc"
	"github.com/emil/quickcast/internal/search"
)

func printUsage() {
	fmt.Print(`Quickcast - Raycast for NixOS / Hyprland

Usage:
  quickcast <command> [arguments]

Commands:
  toggle            Toggle the Quickcast launcher visibility
  show              Show the launcher window
  hide              Hide the launcher window
  ping              Check if quickcast daemon is running
  search <query>    Query items from the daemon and print results
  run <query>       Search and immediately run the top matched item's primary action
  help              Display this help message
`)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	client := ipc.NewClient()

	switch cmd {
	case "toggle":
		resp, err := client.Send(ipc.Request{Action: "toggle"})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if !resp.Success {
			fmt.Fprintf(os.Stderr, "Failed: %s\n", resp.Error)
			os.Exit(1)
		}

	case "show":
		resp, err := client.Send(ipc.Request{Action: "show"})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if !resp.Success {
			fmt.Fprintf(os.Stderr, "Failed: %s\n", resp.Error)
			os.Exit(1)
		}

	case "hide":
		resp, err := client.Send(ipc.Request{Action: "hide"})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if !resp.Success {
			fmt.Fprintf(os.Stderr, "Failed: %s\n", resp.Error)
			os.Exit(1)
		}

	case "ping":
		resp, err := client.Send(ipc.Request{Action: "ping"})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Quickcast daemon is NOT running (%v)\n", err)
			os.Exit(1)
		}
		fmt.Printf("Quickcast daemon alive: %v\n", resp.Data)

	case "search":
		query := strings.Join(os.Args[2:], " ")
		resp, err := client.Send(ipc.Request{
			Action: "search",
			Params: &ipc.RequestParams{Query: query},
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		pretty, _ := json.MarshalIndent(resp.Data, "", "  ")
		fmt.Println(string(pretty))

	case "run":
		query := strings.Join(os.Args[2:], " ")
		if query == "" {
			fmt.Fprintln(os.Stderr, "Error: query required for run")
			os.Exit(1)
		}

		resp, err := client.Send(ipc.Request{
			Action: "search",
			Params: &ipc.RequestParams{Query: query},
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		dataBytes, _ := json.Marshal(resp.Data)
		var items []search.Item
		if err := json.Unmarshal(dataBytes, &items); err != nil || len(items) == 0 {
			fmt.Printf("No match found for %q\n", query)
			os.Exit(1)
		}

		topItem := items[0]
		if len(topItem.Actions) == 0 {
			fmt.Printf("No action available for %s\n", topItem.Title)
			os.Exit(1)
		}

		act := topItem.Actions[0]
		fmt.Printf("⚡ Executing %s: %s\n", topItem.Title, act.Title)

		execResp, err := client.Send(ipc.Request{
			Action: "execute",
			Params: &ipc.RequestParams{
				ItemID: topItem.ID,
				Action: &act,
			},
		})
		if err != nil || !execResp.Success {
			fmt.Fprintf(os.Stderr, "Execution failed: %v\n", err)
			os.Exit(1)
		}

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}
