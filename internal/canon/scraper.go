package canon

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var whitespaceRun = regexp.MustCompile(`\s+`)

// Page represents a scraped documentation page.
type Page struct {
	URL     string
	Title   string
	Content string
}

// Scraper fetches and extracts content from flox.dev.
type Scraper struct {
	client  *http.Client
	visited map[string]bool
}

func NewScraper() *Scraper {
	// Disable keep-alive to free TLS memory between requests.
	transport := &http.Transport{
		DisableKeepAlives: true,
	}
	return &Scraper{
		client:  &http.Client{Timeout: 30 * time.Second, Transport: transport},
		visited: make(map[string]bool),
	}
}

// ScrapeRecursive crawls a base URL and follows links under /docs/ and /blog/.
// ScrapeRecursive crawls a base URL and follows links under /docs/ and /blog/.
// Returns pages collected up to maxPages.
func (s *Scraper) ScrapeRecursive(baseURL string, maxPages int) ([]Page, error) {
	var pages []Page
	s.Crawl(baseURL, maxPages, func(p Page) {
		pages = append(pages, p)
	})
	return pages, nil
}

// Crawl visits pages and calls onPage for each one. The Page is only valid
// during the callback — it is not retained after the callback returns.
func (s *Scraper) Crawl(baseURL string, maxPages int, onPage func(Page)) {
	queue := []string{baseURL}
	count := 0

	for len(queue) > 0 && count < maxPages {
		u := queue[0]
		queue = queue[1:]

		if s.visited[u] {
			continue
		}
		s.visited[u] = true

		page, links, err := s.fetchPage(u)
		if err != nil {
			fmt.Printf("  skip %s: %v\n", u, err)
			continue
		}
		if page.Content != "" {
			onPage(*page)
			count++
		}

		for _, link := range links {
			if !s.visited[link] && (strings.Contains(link, "/docs/") || strings.Contains(link, "/blog/")) {
				queue = append(queue, link)
			}
		}
	}
}

func (s *Scraper) fetchPage(pageURL string) (*Page, []string, error) {
	resp, err := s.client.Get(pageURL)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Use the streaming tokenizer — never holds full DOM in memory.
	title, content, links := tokenize(resp.Body, pageURL)

	// Cap content length.
	if len(content) > 8000 {
		content = content[:8000]
	}

	return &Page{URL: pageURL, Title: title, Content: content}, links, nil
}

// tokenize uses html.Tokenizer (streaming) to extract title, text content, and links
// without building a DOM tree.
func tokenize(r io.Reader, baseURL string) (title, content string, links []string) {
	// Limit reader to 256KB.
	z := html.NewTokenizer(io.LimitReader(r, 256*1024))

	var (
		titleBuf   strings.Builder
		contentBuf strings.Builder
		inTitle    bool
		inMain     bool
		inBody     bool
		skipDepth  int // depth inside script/style/nav/footer/svg
		seenLinks  = make(map[string]bool)
	)

	skipTags := map[string]bool{
		"script": true, "style": true, "nav": true,
		"footer": true, "svg": true, "noscript": true,
	}

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			goto done

		case html.StartTagToken, html.SelfClosingTagToken:
			tn, hasAttr := z.TagName()
			tag := string(tn)

			if skipTags[tag] {
				skipDepth++
				continue
			}

			switch tag {
			case "title":
				inTitle = true
			case "main", "article":
				inMain = true
			case "body":
				inBody = true
			}

			// Extract href from <a> tags.
			if tag == "a" && hasAttr {
				for {
					key, val, more := z.TagAttr()
					if string(key) == "href" {
						link := resolveLink(string(val), baseURL)
						if link != "" && !seenLinks[link] {
							seenLinks[link] = true
							links = append(links, link)
						}
					}
					if !more {
						break
					}
				}
			}

		case html.EndTagToken:
			tn, _ := z.TagName()
			tag := string(tn)

			if skipTags[tag] && skipDepth > 0 {
				skipDepth--
				continue
			}

			switch tag {
			case "title":
				inTitle = false
			}

		case html.TextToken:
			if skipDepth > 0 {
				continue
			}
			text := strings.TrimSpace(string(z.Text()))
			if text == "" {
				continue
			}

			if inTitle {
				titleBuf.WriteString(text)
			}
			// Prefer main/article content over full body.
			if inMain || inBody {
				contentBuf.WriteString(text)
				contentBuf.WriteString(" ")
			}
		}
	}

done:
	title = strings.TrimSpace(titleBuf.String())
	content = cleanText(contentBuf.String())
	return
}

func resolveLink(href, baseURL string) string {
	if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "javascript:") || strings.HasPrefix(href, "mailto:") {
		return ""
	}

	var full string
	if strings.HasPrefix(href, "http") {
		full = href
	} else if strings.HasPrefix(href, "/") {
		full = "https://flox.dev" + href
	} else {
		return ""
	}

	// Strip fragment.
	if idx := strings.Index(full, "#"); idx >= 0 {
		full = full[:idx]
	}

	if !strings.Contains(full, "flox.dev") {
		return ""
	}
	return full
}

func cleanText(s string) string {
	return strings.TrimSpace(whitespaceRun.ReplaceAllString(s, " "))
}
