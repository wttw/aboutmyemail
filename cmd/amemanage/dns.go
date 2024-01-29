package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"net"
	"strings"
	"time"
)

type DnsCmd struct {
	Hostname string `help:"Check DNS for this hostname" required:""`
}

func (a *DnsCmd) Run(globals *Globals) error {
	resolver := net.Resolver{
		PreferGo:     true,
		StrictErrors: true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	res, err := resolver.LookupCNAME(ctx, "cnametest.aboutmy.email")
	if err != nil {
		fatalDNS("Test CNAME lookup failed: %s", err)
	}
	if res != "aboutmy.email." {
		fatal("Test CNAME lookup got wrong answer: %s", res)
	}

	res, err = resolver.LookupCNAME(ctx, a.Hostname)
	if err != nil {
		fatalDNS("CNAME lookup failed: %s", err)
	}
	if hostEq(res, "whitelabel.aboutmy.email") {
		printSuccess(globals, "CNAME OK")
		return nil
	}
	if !hostEq(res, a.Hostname) {
		fatal("CNAME points to %s, should be whitelabel.aboutmy.email", res)
	}
	fatal("No CNAME found for %s", a.Hostname)
	return nil
}

func hostEq(a, b string) bool {
	return strings.EqualFold(strings.TrimSuffix(a, "."), strings.TrimSuffix(b, "."))
}

func fatalDNS(msg string, err error) {
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		blue := color.New(color.FgHiBlue).SprintFunc()
		_, _ = fmt.Fprintf(color.Output, "Querying:    %s\n", blue(dnsErr.Name))
		_, _ = fmt.Fprintf(color.Output, "Nameserver:  %s\n", blue(dnsErr.Server))
		nxdomain := "no"
		if dnsErr.IsNotFound {
			nxdomain = "yes"
		}
		_, _ = fmt.Fprintf(color.Output, "NXDOMAIN:    %s\n", blue(nxdomain))
		fatal(msg, dnsErr.Error())
	}
	fatal(msg, err)
}
