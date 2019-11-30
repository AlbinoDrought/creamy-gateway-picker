package remote

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req"
)

type sensemillaClient struct {
	host     string
	username string
	password string
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
		"login":        "Sign In",
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

func (client *sensemillaClient) gateways() (*goquery.Document, error) {
	return client.fetchOrLogin(func() (*goquery.Document, error) {
		path, err := client.path("/status_gateways.php")
		if err != nil {
			return nil, err
		}

		result, err := req.Get(path)
		if err != nil {
			return nil, err
		}

		resp := result.Response()
		if resp == nil {
			return nil, errors.New("unexpected nil response during ListGateways")
		}

		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("unexpected status code %d when getting gateways", resp.StatusCode)
		}

		return goquery.NewDocumentFromReader(resp.Body)
	})
}

func (client *sensemillaClient) ListGateways() ([]Gateway, error) {
	doc, err := client.gateways()
	if err != nil {
		return nil, err
	}

	gatewayRows := doc.Find(".table-responsive .table tbody tr")

	gateways := make([]Gateway, gatewayRows.Length())

	gatewayRows.Each(func(i int, s *goquery.Selection) {
		/*
			1 Name
			2 Gateway
			3 Monitor
			4 RTT
			5 RTTsd
			6 Loss
			7 Status
			8 Description
		*/
		gateway := &sensemillaGateway{
			name:        strings.TrimSpace(s.Find("td:nth-child(1)").Text()),
			gateway:     strings.TrimSpace(s.Find("td:nth-child(2)").Text()),
			monitor:     strings.TrimSpace(s.Find("td:nth-child(3)").Text()),
			rtt:         strings.TrimSpace(s.Find("td:nth-child(4)").Text()),
			rttsd:       strings.TrimSpace(s.Find("td:nth-child(5)").Text()),
			loss:        strings.TrimSpace(s.Find("td:nth-child(6)").Text()),
			status:      strings.TrimSpace(s.Find("td:nth-child(7)").Text()),
			description: strings.TrimSpace(s.Find("td:nth-child(8)").Text()),
		}

		gateway.online = s.Find("td:nth-child(7)").HasClass("bg-success")

		gateways[i] = gateway
	})

	return gateways, nil
}

func (client *sensemillaClient) firewallRules(iface string) (*goquery.Document, error) {
	return client.fetchOrLogin(func() (*goquery.Document, error) {
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
}

func (client *sensemillaClient) ListRules(iface string) ([]FirewallRule, error) {
	doc, err := client.firewallRules(iface)
	if err != nil {
		return nil, err
	}

	ruleRows := doc.Find("#ruletable tbody tr")

	rules := make([]FirewallRule, ruleRows.Length())

	ruleRows.Each(func(i int, s *goquery.Selection) {
		/*
				1 <!-- checkbox -->
				2 <!-- status icons -->
				3 States
				4 Protocol
				5 Source
				6 Port
				7 Destination
				8 Port
				9 Gateway
			 10 Queue
			 11	Schedule
			 12	Description
			 13	Actions
		*/

		id, _ := s.Find("input[type=\"checkbox\"]").Attr("value")

		rule := &sensemillaFirewallRule{
			id:          id,
			iface:       iface,
			source:      strings.TrimSpace(s.Find("td:nth-child(5)").Text()),
			destination: strings.TrimSpace(s.Find("td:nth-child(7)").Text()),
			gateway:     strings.TrimSpace(s.Find("td:nth-child(9)").Text()),
			description: strings.TrimSpace(s.Find("td:nth-child(12)").Text()),

			client: client,
		}
		rules[i] = rule
	})

	return rules, nil
}

func (client *sensemillaClient) applyChanges(document *goquery.Document, sendRequest func(req.Param) error) error {
	form := document.Find(".alert-warning form.pull-right")
	if form.Length() <= 0 {
		return errors.New("unable to find Apply Changes form")
	}

	csrf, csrfFound := form.Find("input[name=\"__csrf_magic\"]").Attr("value")
	if !csrfFound {
		return errors.New("could not find CSRF input value")
	}

	return sendRequest(req.Param{
		"__csrf_magic": csrf,
		"apply":        "Apply Changes",
	})
}

func (client *sensemillaClient) applyChangesFirewallRules(doc *goquery.Document, iface string) error {
	return client.applyChanges(doc, func(params req.Param) error {
		ifacePath, err := client.path("/firewall_rules.php")
		if err != nil {
			return err
		}

		result, err := req.Post(ifacePath, req.QueryParam{"if": iface}, params)
		if err != nil {
			return err
		}

		resp := result.Response()
		if resp == nil {
			return errors.New("unexpected nil response during apply changes")
		}

		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("unexpected status code %d when applying changes for iface %v", resp.StatusCode, iface)
		}

		return nil
	})
}

func (client *sensemillaClient) AddRule(iface, source, destination, gateway, description string) (FirewallRule, error) {
	rules, err := client.ListRules(iface)
	if err != nil {
		return nil, err
	}

	// hacky logic:
	// shove the new rule after the first rule that has a gateway other than *
	// should probably switch to using a description dork or something instead...
	afterID := ""
	for _, rule := range rules {
		senseRule, ok := rule.(*sensemillaFirewallRule)
		if !ok {
			continue
		}

		if senseRule.Gateway() != "*" {
			break
		}

		afterID = senseRule.id
	}

	ifacePath, err := client.path("/firewall_rules_edit.php")
	if err != nil {
		return nil, err
	}

	doc, err := client.fetchOrLogin(func() (*goquery.Document, error) {
		result, err := req.Get(ifacePath, req.QueryParam{"if": iface})
		if err != nil {
			return nil, err
		}

		resp := result.Response()
		if resp == nil {
			return nil, errors.New("unexpected nil response during AddRule")
		}

		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("unexpected status code %d when visiting add rule page for iface %v", resp.StatusCode, iface)
		}

		return goquery.NewDocumentFromReader(resp.Body)
	})

	if err != nil {
		return nil, err
	}

	csrf, csrfFound := doc.Find("input[name=\"__csrf_magic\"]").Attr("value")
	if !csrfFound {
		return nil, errors.New("could not find CSRF input value")
	}

	var srcParam req.Param
	var destParam req.Param

	if source == "*" {
		srcParam = req.Param{
			"srctype": "any",
		}
	} else {
		srcParam = req.Param{
			"srctype": "single",
			"src":     source,
		}
	}

	if destination == "*" {
		destParam = req.Param{
			"dsttype": "any",
		}
	} else {
		destParam = req.Param{
			"dsttype": "single",
			"dst":     destination,
		}
	}

	result, err := req.Post(ifacePath, srcParam, destParam, req.Param{
		"__csrf_magic": csrf,
		"interface":    iface,
		"descr":        description,
		"gateway":      gateway,
		"after":        afterID,

		// guff:
		"type":               "pass",
		"ipprotocol":         "inet",
		"proto":              "any",
		"icmptype[]":         "any",
		"dscp":               "",
		"tag":                "",
		"tagged":             "",
		"max":                "",
		"max-src-nodes":      "",
		"max-src-conn":       "",
		"max-src-states":     "",
		"max-src-conn-rate":  "",
		"max-src-conn-rates": "",
		"statetimeout":       "",
		"statetype":          "keep state",
		"vlanprio":           "",
		"vlanprioset":        "",
		"sched":              "",
		"dnpipe":             "",
		"pdnpipe":            "",
		"ackqueue":           "",
		"defaultqueue":       "",
		"ruleid":             "",
		"save":               "Save",
	})

	resp := result.Response()
	if resp == nil {
		return nil, errors.New("unexpected nil response during AddRule")
	}

	defer resp.Body.Close()
	if resp.StatusCode != 302 && resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code %d when adding rule for iface %v", resp.StatusCode, iface)
	}

	doc, err = client.firewallRules(iface)
	if err != nil {
		return nil, err
	}

	err = client.applyChangesFirewallRules(doc, iface)
	if err != nil {
		return nil, err
	}

	rules, err = client.ListRules(iface)
	if err != nil {
		return nil, err
	}

	for _, rule := range rules {
		if rule.Source() == source && rule.Gateway() == gateway && rule.Destination() == destination && rule.Description() == description {
			return rule, nil
		}
	}

	return nil, errors.New("unable to find created rule")
}

func (client *sensemillaClient) deleteRule(iface string, id string) error {
	doc, err := client.firewallRules(iface)
	if err != nil {
		return err
	}

	csrf, csrfFound := doc.Find("input[name=\"__csrf_magic\"]").Attr("value")
	if !csrfFound {
		return errors.New("could not find CSRF input value")
	}

	ifacePath, err := client.path("/firewall_rules.php")
	if err != nil {
		return err
	}

	result, err := req.Post(ifacePath, req.QueryParam{"if": iface}, req.Param{
		"__csrf_magic": csrf,
		"act":          "del",
		"if":           iface,
		"id":           id,
	})
	if err != nil {
		return err
	}

	resp := result.Response()
	if resp == nil {
		return errors.New("unexpected nil response during deleteRule")
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 302 {
		return fmt.Errorf("unexpected status code %d when deleting rule %v for iface %v", resp.StatusCode, id, iface)
	}

	doc, err = client.firewallRules(iface)
	if err != nil {
		return err
	}

	err = client.applyChangesFirewallRules(doc, iface)
	if err != nil {
		return err
	}

	return nil
}

// NewSensemillaClient returns a new remote.Client compatible with
// Sensemilla-ish Web UI
func NewSensemillaClient(host, username, password string) Client {
	return &sensemillaClient{
		host,
		username,
		password,
	}
}
