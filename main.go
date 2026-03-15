package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var apiKey string

const baseURL = "https://maps.googleapis.com/maps/api"

// --- Input/Output types ---

type GeocodeInput struct {
	Address string `json:"address" jsonschema:"Address to geocode"`
}

type ReverseGeocodeInput struct {
	Latitude  float64 `json:"latitude" jsonschema:"Latitude coordinate"`
	Longitude float64 `json:"longitude" jsonschema:"Longitude coordinate"`
}

type DirectionsInput struct {
	Origin      string `json:"origin" jsonschema:"Origin address or place"`
	Destination string `json:"destination" jsonschema:"Destination address or place"`
	Mode        string `json:"mode,omitempty" jsonschema:"Travel mode: driving (default), walking, bicycling, or transit"`
}

type DistanceMatrixInput struct {
	Origins      []string `json:"origins" jsonschema:"List of origin addresses"`
	Destinations []string `json:"destinations" jsonschema:"List of destination addresses"`
	Mode         string   `json:"mode,omitempty" jsonschema:"Travel mode: driving (default), walking, bicycling, or transit"`
}

type SearchPlacesInput struct {
	Query string `json:"query" jsonschema:"Search query for places (e.g. pizza near Times Square)"`
}

type Empty struct{}

// --- Google Maps API response types ---

type geocodeResponse struct {
	Results []struct {
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
		PlaceID string `json:"place_id"`
	} `json:"results"`
	Status string `json:"status"`
}

type directionsResponse struct {
	Routes []struct {
		Summary string `json:"summary"`
		Legs    []struct {
			Distance struct {
				Text string `json:"text"`
			} `json:"distance"`
			Duration struct {
				Text string `json:"text"`
			} `json:"duration"`
			StartAddress string `json:"start_address"`
			EndAddress   string `json:"end_address"`
			Steps        []struct {
				HTMLInstructions string `json:"html_instructions"`
				Distance         struct {
					Text string `json:"text"`
				} `json:"distance"`
				Duration struct {
					Text string `json:"text"`
				} `json:"duration"`
				TravelMode string `json:"travel_mode"`
			} `json:"steps"`
		} `json:"legs"`
	} `json:"routes"`
	Status string `json:"status"`
}

type distanceMatrixResponse struct {
	OriginAddresses      []string `json:"origin_addresses"`
	DestinationAddresses []string `json:"destination_addresses"`
	Rows                 []struct {
		Elements []struct {
			Status   string `json:"status"`
			Distance struct {
				Text string `json:"text"`
			} `json:"distance"`
			Duration struct {
				Text string `json:"text"`
			} `json:"duration"`
		} `json:"elements"`
	} `json:"rows"`
	Status string `json:"status"`
}

type placesResponse struct {
	Results []struct {
		Name             string  `json:"name"`
		FormattedAddress string  `json:"formatted_address"`
		Rating           float64 `json:"rating"`
		UserRatingsTotal int     `json:"user_ratings_total"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
		PlaceID string `json:"place_id"`
		Types   []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}

// --- Helpers ---

func mapsGet(endpoint string, params url.Values) ([]byte, error) {
	params.Set("key", apiKey)
	resp, err := http.Get(baseURL + endpoint + "?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	return body, nil
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}
}

func errorResult(msg string) (*mcp.CallToolResult, Empty, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
		IsError: true,
	}, Empty{}, nil
}

func stripHTML(s string) string {
	var out strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			out.WriteRune(r)
		}
	}
	return out.String()
}

// --- Tool handlers ---

func handleGeocode(ctx context.Context, req *mcp.CallToolRequest, input GeocodeInput) (*mcp.CallToolResult, Empty, error) {
	params := url.Values{"address": {input.Address}}
	body, err := mapsGet("/geocode/json", params)
	if err != nil {
		return errorResult(fmt.Sprintf("API error: %v", err))
	}

	var resp geocodeResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return errorResult(fmt.Sprintf("Parse error: %v", err))
	}
	if resp.Status != "OK" {
		return errorResult(fmt.Sprintf("Geocoding failed: %s", resp.Status))
	}

	var sb strings.Builder
	for _, r := range resp.Results {
		fmt.Fprintf(&sb, "Address: %s\n", r.FormattedAddress)
		fmt.Fprintf(&sb, "Latitude: %f\n", r.Geometry.Location.Lat)
		fmt.Fprintf(&sb, "Longitude: %f\n", r.Geometry.Location.Lng)
		fmt.Fprintf(&sb, "Place ID: %s\n\n", r.PlaceID)
	}
	return textResult(sb.String()), Empty{}, nil
}

func handleReverseGeocode(ctx context.Context, req *mcp.CallToolRequest, input ReverseGeocodeInput) (*mcp.CallToolResult, Empty, error) {
	latlng := fmt.Sprintf("%f,%f", input.Latitude, input.Longitude)
	params := url.Values{"latlng": {latlng}}
	body, err := mapsGet("/geocode/json", params)
	if err != nil {
		return errorResult(fmt.Sprintf("API error: %v", err))
	}

	var resp geocodeResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return errorResult(fmt.Sprintf("Parse error: %v", err))
	}
	if resp.Status != "OK" {
		return errorResult(fmt.Sprintf("Reverse geocoding failed: %s", resp.Status))
	}

	var sb strings.Builder
	for i, r := range resp.Results {
		if i >= 3 {
			break
		}
		fmt.Fprintf(&sb, "Address: %s\n", r.FormattedAddress)
		fmt.Fprintf(&sb, "Place ID: %s\n\n", r.PlaceID)
	}
	return textResult(sb.String()), Empty{}, nil
}

func handleDirections(ctx context.Context, req *mcp.CallToolRequest, input DirectionsInput) (*mcp.CallToolResult, Empty, error) {
	params := url.Values{
		"origin":      {input.Origin},
		"destination": {input.Destination},
	}
	if input.Mode != "" {
		params.Set("mode", input.Mode)
	}

	body, err := mapsGet("/directions/json", params)
	if err != nil {
		return errorResult(fmt.Sprintf("API error: %v", err))
	}

	var resp directionsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return errorResult(fmt.Sprintf("Parse error: %v", err))
	}
	if resp.Status != "OK" {
		return errorResult(fmt.Sprintf("Directions failed: %s", resp.Status))
	}

	var sb strings.Builder
	for i, route := range resp.Routes {
		if i > 0 {
			sb.WriteString("---\n")
		}
		fmt.Fprintf(&sb, "Route: %s\n", route.Summary)
		for _, leg := range route.Legs {
			fmt.Fprintf(&sb, "From: %s\n", leg.StartAddress)
			fmt.Fprintf(&sb, "To: %s\n", leg.EndAddress)
			fmt.Fprintf(&sb, "Distance: %s\n", leg.Distance.Text)
			fmt.Fprintf(&sb, "Duration: %s\n\n", leg.Duration.Text)
			fmt.Fprintf(&sb, "Steps:\n")
			for j, step := range leg.Steps {
				fmt.Fprintf(&sb, "  %d. %s (%s, %s)\n", j+1, stripHTML(step.HTMLInstructions), step.Distance.Text, step.Duration.Text)
			}
			sb.WriteString("\n")
		}
	}
	return textResult(sb.String()), Empty{}, nil
}

func handleDistanceMatrix(ctx context.Context, req *mcp.CallToolRequest, input DistanceMatrixInput) (*mcp.CallToolResult, Empty, error) {
	params := url.Values{
		"origins":      {strings.Join(input.Origins, "|")},
		"destinations": {strings.Join(input.Destinations, "|")},
	}
	if input.Mode != "" {
		params.Set("mode", input.Mode)
	}

	body, err := mapsGet("/distancematrix/json", params)
	if err != nil {
		return errorResult(fmt.Sprintf("API error: %v", err))
	}

	var resp distanceMatrixResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return errorResult(fmt.Sprintf("Parse error: %v", err))
	}
	if resp.Status != "OK" {
		return errorResult(fmt.Sprintf("Distance matrix failed: %s", resp.Status))
	}

	var sb strings.Builder
	for i, row := range resp.Rows {
		origin := "Unknown"
		if i < len(resp.OriginAddresses) {
			origin = resp.OriginAddresses[i]
		}
		for j, elem := range row.Elements {
			dest := "Unknown"
			if j < len(resp.DestinationAddresses) {
				dest = resp.DestinationAddresses[j]
			}
			fmt.Fprintf(&sb, "%s → %s\n", origin, dest)
			if elem.Status == "OK" {
				fmt.Fprintf(&sb, "  Distance: %s\n", elem.Distance.Text)
				fmt.Fprintf(&sb, "  Duration: %s\n\n", elem.Duration.Text)
			} else {
				fmt.Fprintf(&sb, "  Status: %s\n\n", elem.Status)
			}
		}
	}
	return textResult(sb.String()), Empty{}, nil
}

func handleSearchPlaces(ctx context.Context, req *mcp.CallToolRequest, input SearchPlacesInput) (*mcp.CallToolResult, Empty, error) {
	params := url.Values{"query": {input.Query}}
	body, err := mapsGet("/place/textsearch/json", params)
	if err != nil {
		return errorResult(fmt.Sprintf("API error: %v", err))
	}

	var resp placesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return errorResult(fmt.Sprintf("Parse error: %v", err))
	}
	if resp.Status != "OK" {
		return errorResult(fmt.Sprintf("Places search failed: %s", resp.Status))
	}

	var sb strings.Builder
	for i, p := range resp.Results {
		if i >= 10 {
			break
		}
		fmt.Fprintf(&sb, "%d. %s\n", i+1, p.Name)
		fmt.Fprintf(&sb, "   Address: %s\n", p.FormattedAddress)
		if p.Rating > 0 {
			fmt.Fprintf(&sb, "   Rating: %.1f (%d reviews)\n", p.Rating, p.UserRatingsTotal)
		}
		fmt.Fprintf(&sb, "   Location: %f, %f\n", p.Geometry.Location.Lat, p.Geometry.Location.Lng)
		fmt.Fprintf(&sb, "   Place ID: %s\n", p.PlaceID)
		if len(p.Types) > 0 {
			fmt.Fprintf(&sb, "   Types: %s\n", strings.Join(p.Types, ", "))
		}
		sb.WriteString("\n")
	}
	return textResult(sb.String()), Empty{}, nil
}

func main() {
	apiKey = os.Getenv("GOOGLE_MAPS_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_MAPS_API_KEY environment variable is required")
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "google-maps",
		Version: "1.0.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "maps_geocode",
		Description: "Convert an address to latitude/longitude coordinates",
	}, handleGeocode)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "maps_reverse_geocode",
		Description: "Convert latitude/longitude coordinates to an address",
	}, handleReverseGeocode)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "maps_directions",
		Description: "Get directions between two locations with step-by-step navigation",
	}, handleDirections)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "maps_distance_matrix",
		Description: "Calculate distances and travel times between multiple origins and destinations",
	}, handleDistanceMatrix)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "maps_search_places",
		Description: "Search for places by text query (e.g. 'coffee shops near Central Park')",
	}, handleSearchPlaces)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
