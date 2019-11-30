package remote

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req"
)

type sensemillaFirewallRule struct {
	source      string
	destination string
	gateway     string
	description string
}

func (rule *sensemillaFirewallRule) Source() string {
	return rule.source
}

func (rule *sensemillaFirewallRule) Destination() string {
	return rule.destination
}

func (rule *sensemillaFirewallRule) Gateway() string {
	return rule.gateway
}

func (rule *sensemillaFirewallRule) Description() string {
	return rule.description
}

type sensemillaClient struct {
	host     string
	username string
	password string
	// referrer string
}

func (client *sensemillaClient) path(path string) (string, error) {
	url, err := url.Parse(client.host)
	if err != nil {
		return "", err
	}
	url.Path = path
	return url.String(), nil
}

func (client *sensemillaClient) loggedOut(document *goquery.Document) bool {
	return document.Find("form.login").Length() > 0
}

func (client *sensemillaClient) loginIfRequired(document *goquery.Document) error {
	if !client.loggedOut(document) {
		return nil
	}

	csrf, csrfFound := document.Find("input[name=\"__csrf_magic\"]").Attr("value")
	if !csrfFound {
		return errors.New("could not find CSRF input value")
	}

	result, err := req.Post(client.host, req.Param{
		"__csrf_magic": csrf,
		"usernamefld":  client.username,
		"passwordfld":  client.password,
		"login":        "Sign+In",
	})

	if err != nil {
		return err
	}

	resp := result.Response()
	if resp == nil {
		return errors.New("unexpected nil response during login")
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status code %d when logging in", resp.StatusCode)
	}

	document, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	if !client.loggedOut(document) {
		return nil
	}

	return errors.New("login failed")
}

func (client *sensemillaClient) fetchOrLogin(fetch func() (*goquery.Document, error)) (*goquery.Document, error) {
	document, err := fetch()
	if err != nil {
		return nil, err
	}

	if client.loggedOut(document) {
		err = client.loginIfRequired(document)
		if err != nil {
			return nil, err
		}
		document, err = fetch()
	}

	return document, err
}

func (client *sensemillaClient) ListRules(iface string) ([]FirewallRule, error) {
	doc, err := client.fetchOrLogin(func() (*goquery.Document, error) {
		ifacePath, err := client.path("/firewall_rules.php")
		if err != nil {
			return nil, err
		}

		result, err := req.Get(ifacePath, req.QueryParam{"if": iface})
		if err != nil {
			return nil, err
		}

		resp := result.Response()
		if resp == nil {
			return nil, errors.New("unexpected nil response during ListRules")
		}

		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("unexpected status code %d when fetching rules for iface %v", resp.StatusCode, iface)
		}

		return goquery.NewDocumentFromReader(resp.Body)
	})

	if err != nil {
		return nil, err
	}

	ruleRows := doc.Find("#ruletable tbody tr")

	rules := make([]FirewallRule, ruleRows.Length())

	ruleRows.Each(func(i int, s *goquery.Selection) {
		/*
				0 <!-- checkbox -->
				1 <!-- status icons -->
				2 States
				3 Protocol
				4 Source
				5 Port
				6 Destination
				7 Port
				8 Gateway
				9 Queue
			 10	Schedule
			 11	Description
			 12	Actions
		*/

		rule := &sensemillaFirewallRule{
			source:      strings.TrimSpace(s.Find("td:nth-child(5)").Text()),
			destination: strings.TrimSpace(s.Find("td:nth-child(7)").Text()),
			gateway:     strings.TrimSpace(s.Find("td:nth-child(9)").Text()),
			description: strings.TrimSpace(s.Find("td:nth-child(12)").Text()),
		}
		rules[i] = rule
	})

	return rules, nil
}

func NewSensemillaClient(host, username, password string) Client {
	return &sensemillaClient{
		host,
		username,
		password,
		// referrer,
	}
}
