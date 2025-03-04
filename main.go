package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Availability struct {
	Storage      string       `json:"storage"`
	Memory       string       `json:"memory"`
	RessourcesID string       `json:"fqn"`
	Datacenters  []Datacenter `json:"datacenters"`
}

type Datacenter struct {
	Availability string `json:"availability"`
	Datacenter   string `json:"datacenter"`
}

type DiscordWebhook struct {
	Content   string         `json:"content,omitempty"`
	Username  string         `json:"username,omitempty"`
	AvatarUrl string         `json:"avatar_url,omitempty"`
	Embeds    []DiscordEmbed `json:"embeds,omitempty"`
}

type DiscordEmbed struct {
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	Color       int          `json:"color"`
	Fields      []EmbedField `json:"fields,omitempty"`
	Footer      *EmbedFooter `json:"footer,omitempty"`
}

type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type EmbedFooter struct {
	Text string `json:"text"`
}

func main() {
	fmt.Println("OVH Stock Checker started, running every hour")

	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	discordMessageContent := os.Getenv("DISCORD_MESSAGE_CONTENT")
	if webhookURL == "" {
		fmt.Println("Warning: DISCORD_WEBHOOK_URL environment variable not set")
		fmt.Println("Notifications will not be sent to Discord")
	} else {
		fmt.Println("Discord webhook URL is set to: " + webhookURL)
	}
	if discordMessageContent != "" {
		fmt.Println("Discord message is set to: " + discordMessageContent)
	}

	loopCount := 1

	fmt.Printf("[Check %d] Running scheduled check at: %s\n", loopCount, time.Now().Format("2006-01-02 15:04:05"))
	checkOvhStocks(webhookURL, discordMessageContent)
	fmt.Printf("[Check %d] Finished at: %s\n", loopCount, time.Now().Format("2006-01-02 15:04:05"))
	loopCount++

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Printf("[Check %d] Running scheduled check at: %s\n", loopCount, time.Now().Format("2006-01-02 15:04:05"))
			checkOvhStocks(webhookURL, discordMessageContent)
			fmt.Printf("[Check %d] Finished at: %s\n", loopCount, time.Now().Format("2006-01-02 15:04:05"))
			loopCount++
		}
	}
}

func checkOvhStocks(webhookURL string, discordMessageContent string) {
	allServersData := make(map[string]Availability)

	checkAndCollectAvailability(&allServersData, "https://www.ovh.com/engine/apiv6/dedicated/server/datacenter/availabilities/?excludeDatacenters=false&planCode=24sk10", "24sk10.ram-32g-ecc-2133.softraid-2x480ssd", "KS-1")
	checkAndCollectAvailability(&allServersData, "https://www.ovh.com/engine/apiv6/dedicated/server/datacenter/availabilities/?excludeDatacenters=false&planCode=24sk20", "24sk20.ram-32g-ecc-2133.softraid-2x450nvme", "KS-2")
	checkAndCollectAvailability(&allServersData, "https://www.ovh.com/engine/apiv6/dedicated/server/datacenter/availabilities/?excludeDatacenters=false&planCode=24sk30", "24sk30.ram-32g-ecc-2133.softraid-2x480ssd", "KS-3")
	checkAndCollectAvailability(&allServersData, "https://www.ovh.com/engine/apiv6/dedicated/server/datacenter/availabilities/?excludeDatacenters=false&planCode=24sk40", "24sk40.ram-32g-ecc-2133.softraid-2x450nvme", "KS-4")

	atLeastOneServerAvailable := false
	for _, availability := range allServersData {
		for _, dc := range availability.Datacenters {
			if dc.Availability != "unavailable" {
				atLeastOneServerAvailable = true
				break
			}
		}
		if atLeastOneServerAvailable {
			break
		}
	}

	if atLeastOneServerAvailable {
		sendSummaryToDiscord(webhookURL, discordMessageContent, allServersData)
		fmt.Println("At least one server is available. Sent summary to Discord.")
	} else {
		fmt.Println("No servers are available.")
	}
}

func checkAndCollectAvailability(allServersData *map[string]Availability, url string, resources string, serverName string) {
	// Make the request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error making request for %s: %v\n", serverName, err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error closing response body for %s: %v\n", serverName, err)
		}
	}(resp.Body)

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response for %s: %v\n", serverName, err)
		return
	}

	// Parse the JSON
	var availabilities []Availability
	if err := json.Unmarshal(body, &availabilities); err != nil {
		fmt.Printf("Error parsing JSON for %s: %v\n", serverName, err)
		return
	}

	// Find the object with correct storage
	for _, avail := range availabilities {
		if strings.ToLower(avail.RessourcesID) == resources {
			(*allServersData)[serverName] = avail
			fmt.Println("Finished checking availability for " + serverName)
			return
		}
	}

	fmt.Println("No availability found for server: " + serverName)
}

func sendSummaryToDiscord(webhookURL string, discordMessageContent string, allServersData map[string]Availability) {
	var fields []EmbedField

	for serverName, availability := range allServersData {
		serverAvailable := false
		for _, dc := range availability.Datacenters {
			if dc.Availability != "unavailable" {
				serverAvailable = true
				break
			}
		}

		icon := "❌"
		if serverAvailable {
			icon = "✅"
		}

		fields = append(fields, EmbedField{
			Name:   serverName,
			Value:  icon + " [Buy now](https://eco.ovhcloud.com/fr/kimsufi/" + serverName + "/)",
			Inline: true,
		})
	}

	color := 5763719

	webhook := DiscordWebhook{
		Username:  "OVH Server Availability Bot",
		AvatarUrl: "https://s3-symbol-logo.tradingview.com/ovh-groupe--600.png",
		Embeds: []DiscordEmbed{
			{
				Title:  "Dedicated Server Availability Summary",
				Color:  color,
				Fields: fields,
				Footer: &EmbedFooter{
					Text: "Last updated: " + time.Now().Format("2006-01-02 15:04:05"),
				},
			},
		},
		Content: discordMessageContent,
	}

	jsonData, err := json.Marshal(webhook)
	if err != nil {
		fmt.Printf("Error creating JSON payload: %v\n", err)
		return
	}

	// For debugging
	fmt.Printf("Sending payload: %s\n", string(jsonData))

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending webhook: %v\n", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error closing response body when sending discord webhook: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Println("Summary message sent to Discord successfully")
	} else {
		fmt.Printf("Failed to send summary message to Discord. Status code: %d\n", resp.StatusCode)
	}
}
