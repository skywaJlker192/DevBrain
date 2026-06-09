package services

import (
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// FetchMetadata: Извлекает meta-теги со страницы
func FetchMetadata(pageURL string) (title, description string, err error) {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(pageURL)
	if err != nil {
		return extractFromURL(pageURL), "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return extractFromURL(pageURL), "", nil
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return extractFromURL(pageURL), "", nil
	}

	// Ищем <title>
	title = extractFromURL(pageURL)
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			if n.FirstChild != nil {
				title = strings.TrimSpace(n.FirstChild.Data)
			}
		}

		// Ищем OG tags
		if n.Type == html.ElementNode && n.Data == "meta" {
			var property, content string
			for _, attr := range n.Attr {
				if attr.Key == "property" || attr.Key == "name" {
					property = attr.Val
				}
				if attr.Key == "content" {
					content = attr.Val
				}
			}

			if strings.Contains(property, "title") && content != "" {
				title = content
			}
			if strings.Contains(property, "description") && content != "" {
				description = content
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return title, description, nil
}

func extractFromURL(u string) string {
	if parsed, err := url.Parse(u); err == nil {
		host := parsed.Hostname()
		if strings.HasPrefix(host, "www.") {
			host = host[4:]
		}
		return host
	}
	return "Без названия"
}
