package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// Element represents an element with its root, left and right components, and tier.
type Element struct {
	Root  string `json:"root"`
	Left  string `json:"left"`
	Right string `json:"right"`
	Tier  string `json:"tier"`
}

// Scraper scrapes element data from the Little Alchemy 2 wiki and saves it as JSON.
func Scraper() {
	url := "https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)"
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		panic(err)
	}

	html := buf.String()
	var allElements []Element

	// Extract starting elements (Tier 0)
	startingRegex := regexp.MustCompile(`(?is)<span class="mw-headline" id="starting_elements">.*?</span>.*?(<table.*?>.*?</table>)`)
	startingMatch := startingRegex.FindStringSubmatch(html)
	if len(startingMatch) > 1 {
		elements := extractElementsFromTable(startingMatch[1], "0")
		allElements = append(allElements, elements...)
	}

	// Extract tiered elements
	tierRegex := regexp.MustCompile(`(?is)<span class="mw-headline" id="(tier_\d+)_elements">.*?</span>.*?(<table.*?>.*?</table>)`)
	tierMatches := tierRegex.FindAllStringSubmatch(html, -1)

	for _, match := range tierMatches {
		tier := strings.Split(match[1], "_")[1]
		elements := extractElementsFromTable(match[2], tier)
		allElements = append(allElements, elements...)
	}
	
	allElements = append(allElements, Element{
		Root: "Time",
		Left: "",
		Right: "",
		Tier: "0",
	})

	// Save to combinations.json
	data, err := json.MarshalIndent(allElements, "", "  ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("combinations.json", data, 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("Data berhasil disimpan di combinations.json")
}

// extractElementsFromTable extracts elements from an HTML table based on the given tier.
func extractElementsFromTable(tableHTML string, tier string) []Element {
	var results []Element
	rowRegex := regexp.MustCompile(`(?is)<tr>.*?</tr>`)
	rows := rowRegex.FindAllStringSubmatch(tableHTML, -1)

	for _, row := range rows {
		tdRegex := regexp.MustCompile(`(?is)<td.*?>(.*?)</td>`)
		tds := tdRegex.FindAllStringSubmatch(row[0], -1)
		if len(tds) == 0 {
			continue
		}

		root := extractTitle(tds[0][1])
		if root == "" {
			continue
		}

		// Check for combinations (left and right components)
		if len(tds) >= 2 {
			composers := extractComposers(tds[1][1])
			if len(composers) > 0 {
				for _, pair := range composers {
					results = append(results, Element{
						Root:  root,
						Left:  pair[0],
						Right: pair[1],
						Tier:  tier,
					})
				}
				continue
			}
		}

		// If no combination, add as a standalone element
		results = append(results, Element{
			Root:  root,
			Left:  "",
			Right: "",
			Tier:  tier,
		})
	}

	return results
}

// extractTitle extracts the title of an element from HTML.
func extractTitle(html string) string {
	re := regexp.MustCompile(`(?i)<a [^>]*title="([^"]+)"`)
	match := re.FindStringSubmatch(html)
	if len(match) > 1 {
		return cleanText(match[1])
	}
	return ""
}

// cleanText removes extra whitespace from a string.
func cleanText(s string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
}

// extractComposers extracts pairs of components (left and right) from HTML.
func extractComposers(html string) [][2]string {
	var results [][2]string
	liRegex := regexp.MustCompile(`(?is)<li[^>]*>(.*?)</li>`)
	liMatches := liRegex.FindAllStringSubmatch(html, -1)

	for _, li := range liMatches {
		aRegex := regexp.MustCompile(`(?i)<a [^>]*title="([^"]+)"`)
		aMatches := aRegex.FindAllStringSubmatch(li[1], -1)
		if len(aMatches) >= 2 {
			left := cleanText(aMatches[0][1])
			right := cleanText(aMatches[1][1])
			results = append(results, [2]string{left, right})
		}
	}
	return results
}