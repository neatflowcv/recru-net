package saramin

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"

	"github.com/neatflowcv/recru-net/internal/domain"
)

func parseListHTML(data []byte, baseURL string) ([]domain.JobPosting, error) {
	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	anchors := findNodes(doc, func(n *html.Node) bool {
		if n.Type != html.ElementNode || n.Data != "a" {
			return false
		}
		href := getAttr(n, "href")
		return strings.Contains(href, "/zf_user/jobs/relay/view")
	})

	items := make([]domain.JobPosting, 0, len(anchors))
	seen := map[string]struct{}{}
	for _, a := range anchors {
		href := strings.TrimSpace(getAttr(a, "href"))
		if href == "" {
			continue
		}
		if strings.HasPrefix(href, "/") {
			href = strings.TrimSuffix(baseURL, "/") + href
		}
		extID := extractExternalID(href)
		key := extID + "|" + href
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		title := strings.TrimSpace(nodeText(a))
		if title == "" {
			title = strings.TrimSpace(getAttr(a, "title"))
		}

		container := a.Parent
		company := strings.TrimSpace(nodeText(findFirst(container, func(n *html.Node) bool {
			if n.Type != html.ElementNode {
				return false
			}
			class := getAttr(n, "class")
			return strings.Contains(class, "corp_name") || strings.Contains(class, "company")
		})))
		location := strings.TrimSpace(nodeText(findFirst(container, func(n *html.Node) bool {
			if n.Type != html.ElementNode {
				return false
			}
			class := getAttr(n, "class")
			return strings.Contains(class, "job_condition") || strings.Contains(class, "work_place") || strings.Contains(class, "location")
		})))

		items = append(items, domain.JobPosting{
			ExternalID: extID,
			URL:        href,
			Title:      title,
			Company:    company,
			Location:   location,
		})
	}

	return items, nil
}

func parseDetailHTML(data []byte) (domain.JobPosting, error) {
	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return domain.JobPosting{}, err
	}

	companyNode := findFirst(doc, func(n *html.Node) bool {
		if n.Type != html.ElementNode {
			return false
		}
		class := getAttr(n, "class")
		return strings.Contains(class, "company_name") || strings.Contains(class, "corp_name")
	})
	locationNode := findFirst(doc, func(n *html.Node) bool {
		if n.Type != html.ElementNode {
			return false
		}
		class := getAttr(n, "class")
		return strings.Contains(class, "work_place") || strings.Contains(class, "location")
	})
	dateNode := findFirst(doc, func(n *html.Node) bool {
		if n.Type != html.ElementNode {
			return false
		}
		class := getAttr(n, "class")
		return strings.Contains(class, "date") || strings.Contains(class, "deadline")
	})

	postedAt := parseKoreanDate(nodeText(dateNode))
	return domain.JobPosting{
		Company:  strings.TrimSpace(nodeText(companyNode)),
		Location: strings.TrimSpace(nodeText(locationNode)),
		PostedAt: postedAt,
	}, nil
}

func findNodes(root *html.Node, match func(*html.Node) bool) []*html.Node {
	var out []*html.Node
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n == nil {
			return
		}
		if match(n) {
			out = append(out, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	return out
}

func findFirst(root *html.Node, match func(*html.Node) bool) *html.Node {
	if root == nil {
		return nil
	}
	if match(root) {
		return root
	}
	for c := root.FirstChild; c != nil; c = c.NextSibling {
		if found := findFirst(c, match); found != nil {
			return found
		}
	}
	return nil
}

func getAttr(n *html.Node, key string) string {
	if n == nil {
		return ""
	}
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func nodeText(n *html.Node) string {
	if n == nil {
		return ""
	}
	if n.Type == html.TextNode {
		return n.Data
	}
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		b.WriteString(nodeText(c))
		b.WriteString(" ")
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
