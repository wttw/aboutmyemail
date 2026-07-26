package main

import "testing"

// A non-ASCII Return-Path must be used in preference to From, and survive
// intact.
func TestDefaultAddressesPrefersReturnPath(t *testing.T) {
	email := []byte("Return-Path: <grå@grå.org>\r\n" +
		"From: Gøril <gøril@example.com>\r\n" +
		"To: <arnt@grå.org>\r\n" +
		"Subject: test\r\n\r\nbody\r\n")
	from, to := defaultAddresses(email, "", "")
	if from != "grå@grå.org" {
		t.Errorf("from: want %q, got %q", "grå@grå.org", from)
	}
	if to != "arnt@grå.org" {
		t.Errorf("to: want %q, got %q", "arnt@grå.org", to)
	}
}

func TestLocalpartNeedsUTF8(t *testing.T) {
	cases := []struct {
		name  string
		addrs []string
		want  bool
	}{
		{"all ascii", []string{"steve@blighty.com", "arnt@example.com"}, false},
		{"non-ascii domain only", []string{"arnt@grå.org"}, false},
		{"non-ascii localpart", []string{"grå@grå.org"}, true},
		{"non-ascii localpart, ascii domain", []string{"gøril@example.com"}, true},
		{"one of many", []string{"steve@blighty.com", "例子@例子.中国"}, true},
		{"no domain", []string{"उदाहरण"}, true},
		{"empty", nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := localpartNeedsUTF8(c.addrs...); got != c.want {
				t.Errorf("localpartNeedsUTF8(%q) = %v, want %v", c.addrs, got, c.want)
			}
		})
	}
}
