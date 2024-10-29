package xml

import (
	"strings"

	"github.com/antchfx/xmlquery"
)

func ValidateByXpath(msg *string, filterXPath string, filterValue string, xPath string, value string) bool {
	doc, err := xmlquery.Parse(strings.NewReader(*msg))
	if err != nil {
		panic(err)
	}

	if n := xmlquery.FindOne(doc, filterXPath); n != nil {
		id := n.InnerText()
		if !strings.Contains(id, filterValue) {
			// file not part of the test, return true
			return true
		}
	}

	if n := xmlquery.FindOne(doc, xPath); n != nil {
		return n.InnerText() == value
	}

	return false
}
