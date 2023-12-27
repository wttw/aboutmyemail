## About My Email {#faq-about}

AboutMy.email is a tool to analyze an email for technical
good practice. It's currently in an alpha release state, so
expect to see some bugs. If you find any feel free to [file
a bug or mention it on Slack](/help).

## Authentication - SPF {#faq-spf}

SPF queries DNS to check whether an IP address is authorized
to send email that uses a particular return path (or HELO).

It can return a variety of possible statuses, but for our
purposes anything that isn't "pass" is a fail.

## SPF DNS Queries {#faq-spf-dns}

Querying SPF can involve follow pointers to other records,
and from there to other records. The SPF standards limit
the number of DNS related operations to 10, though enforcement
of this by mailbox providers varies.

A void operation is a DNS query for SPF that returns no
results. These are a sign of misconfiguration and if there
are more than two the SPF check will fail.


## Email size {#faq-size}

The size of the email as transmitted. This includes the
headers, plain text and html parts, attachments and any
encoding overhead.

## Email weight {#faq-weight}

The total size of the external images referenced by the email.

The bigger this is, the longer it will take to display the
email. This is a minimum - it only counts images referenced
directly in an `<img>` tag, not srcset or css referenced images.

## Email bloat {#faq-bloat}

An email will usually specify the size of an image in the
`<img>` tag that references it. If the actual image fetched from
the remote server is a different size the mail client will
scale it up or down.

If the image is scaled down by the client
that means the server is sending a larger image than the client
needs. The bloat is the number of bytes that could be saved
by serving a file the same size as the mail client expects.

## Alignment {#faq-aligned}

Alignment is a term that comes from [DMARC](#faq-dmarc).

Loosely it means that two hostnames are in the same "domain",
e.g. www.example.com and mail.example.com are aligned because
they're both in the domain example.com.

More technically it means that the two hostnames share the
same "organizational domain". That's usually the same thing
as your intuitive idea of what a domain is, but it's a
well-defined term. Currently, an organizational domain
is defined
by the [Public Suffix List](https://publicsuffix.org).

Generally when we use the term "aligned" this is what we mean,
but this is what DMARC refers to as "relaxed alignment". The
other type is "strict alignment", which requires the two hostname be identical.

If we describe just one hostname as "aligned" we mean that
it's aligned to the domain of the email address in the From:
header.

## IPv6 {#faq-ipv6}

IPv6 is the "new" type of IP address that was standardised
about 20 years ago, replacing the older IPv4 format.

IPv6 addresses look like `2602:ff16:9:0:1:17b:0:1::`, while
IPv4 addresses look more like `89.233.107.76`.

There's a shortage of IPv4 addresses - we've run out of
newly minted IPv4 addresses completely, and are now just
trading them around.

While it's unlikely there'll be any IPv6-only email servers
for a long time, sending mail via IPv6 rather than IPv4 where
possible can save scarce IPv4 resources (and might deliver with
less delay or throttling in some cases).

IPv6-only clients are more likely to happen, as are IPv6 clients
who only have access to IPv4 content via flaky, overloaded
carrier grade NAT devices. Providing external content, such as
images and landing pages, via IPv6 may make for a better,
faster experience for those clients.

## List-Unsubscribe header {#faq-list-unsubscribe}

The List-Unsubscribe header, defined in [RFC 2369](rfc2369),
contains one or more URLs that a mail client can use to
request a recipient be unsubscribed.

Typically, it will contain `mailto:` or `https` URLs (`http`
URLs may still work, but are obsolete and should be converted
to `https`).

A `mailto:` URL allows a mail client to request unsubscription
non-interactively, withouth the user needing to take any
further action, by sending an email to the given address.

An `https:` URL allows a mail client to open a browser for
the user to visit the given URL. From that page they're
expected to be able to unsubscribe from all email and,
optionally, to use a subscription center to unsubscribe
from subsets of the senders email.

The `https:` URL is also used as part of the protocol for
["one-click" unsubscribes](#faq-list-unsubscribe-post).

## List-Unsubscribe-Post header {#faq-list-unsubscribe-post}

The List-Unsubscribe-Post header, define in [RFC 8058](rfc8058),
is used to enable non-interactive, "one-click", unsubscription.

It contains a fixed value, `List-Unsubscribe=One-Click`, and
signals that a `https:` URL provided in the
[List-Unsubscribe header](#faq-list-unsubscribe) can be used
non-interactively via a http POST to request unsubscription
without any additional user interaction required.

## Non-interactive unsubscribe {#faq-in-app-unsubscribe}

Mailbox providers want to be able to help their users manage
the mailing lists they subscribe to, e.g. 
[Yahoo's Subscription Hub](https://senders.yahooinc.com/subhub/).

They also want to be able to offer user the option to
unsubscribe from email they don't want, for example as
prompt to unsubscribe as an alternative to blocking a
sender in response to the user marking an email as spam.

To provide this user interface the mailbox provider needs
to be able to manage unsubscription on the users behalf,
often described as non-interactive, in-app or one-click
unsubscription.

These can be provided in two ways. The older, deprecated,
approach is a `mailto:` URL in a
[List-Unsubscribe](#faq-list-unsubscribe) header which
will trigger an unsubscribe in response to an email to
a special address. This is asynchronous and doesn't provide
any feedback to the mailbox provider, so isn't a good match
to an interactive, real-time unsubscription interface. The
newer, strongly preferred, approach is the
[List-Unsubscribe-Post header](#faq-list-unsubscribe-post),
which provides a web API for unsubscription.

## In-body unsubscription {#faq-in-body-unsubscribe}

Users expect a link in the footer of an email they can
click on to go to a webpage to unsubscribe from this
sender, or manage their subscriptions. This is the
usual way to comply with [legal requirements](https://wordtothewise.com/2016/04/lets-talk-can-spam/)
around unsubscription. These requirements are usually
_not_ satisified by [List-Unsubscribe headers](#faq-list-unsubscribe), so you need this link in
addition to those, not instead of.

The expected behaviour for this web page is identical to
the expected behaviour of that provided by the
List-Unsubscribe header, so they'll often be identical.
(Adding a parameter to distinguish the two is useful if you
want to be able to report on how a user unsubscribes).

## List-ID header {#faq-list-id}

The List-ID header, defined in [RFC 2369](rfc2369), provides
a way for mail clients to identify a particular mail stream.

It typically contains two parts - an optional human-readable description  and a machine-readable part that
looks something like a hostname.

The machine-readable part is hierarchical, which allows a
mail client to group related mailing lists together easily.

The human-readable part is optional, but if it's missing the
recipient may see the more confusing machine-readable
description in their UI, for instance when managing
subscriptions.

## Reverse DNS {#faq-reverse-dns}

Reverse DNS uses the domain name system to find the hostname
associated with an IP address. This is called reverse DNS as
it's the reverse of the more common case of finding the IP
address associated with a hostname.

Reverse DNS is something that's configured by the owner of
an IP address, which means they can claim to be any hostname
they want. Because of this, most uses of reverse DNS require
[Forward-confirmed reverse DNS](https://wordtothewise.com/2020/02/what-is-fcrdns-and-why-do-we-care/)
(aka FCrDNS, full-circle reverse dns, or round-trip reverse
dns). This allows someone to be sure that the IP address
and hostname are controlled by the same entity.

Reverse DNS is particularly important for IP addresses that
are sending email, as it's the lowest bar for demonstrating
that the IP address belongs to a properly configured server
rather than an infected consumer machine.

It's polite, and demonstrates a level of operational competence,
to have FCrDNS on any IP addresses you're providing public
services from.

## Transport Layer Security (TLS) {#faq-tls}

## Opportunistic TLS {#faq-opportunistic-tls}

SMTP doesn't require TLS encryption, but supports upgrading
an SMTP connection to one encrypted by TLS using the STARTTLS
command. Offering, but not requiring, TLS like this is called
[opportunistic TLS](https://en.wikipedia.org/wiki/Opportunistic_TLS).

We think that any sender of email should ideally use STARTTLS
if the  receiving server offers it.

## Generic Reverse DNS {#faq-generic-reverse-dns}

Hosting providers often mechanically set reverse DNS across
all their IP space, so it will look more like `89-233-107-76.static.example.net` than `mail.example.com`.

This is known as generic reverse DNS, and is often a sign
that the server hasn't been properly set up and maybe shouldn't
be trusted as a source of email.

## Mailbox Provider Expectations {#faq-yahoogle}

[Yahoo](https://senders.yahooinc.com/best-practices/) and
[Google](https://support.google.com/a/answer/14229414) have
announced what behaviour they expect to see from well
behaved bulk senders.

If anything on your [Good Practice](/yahoogle) page is red,
you're probably not complying with their requirements, and
your delivery is likely to suffer.

## But, Steve, you don't follow all these best practices {#faq-but-steve}

And the [cobbler's children go barefoot](https://www.oxfordreference.com/display/10.1093/oi/authority.20110803100502785).

I generally try and set up my services to follow good practices,
but things like reverse DNS are far lower down the list than
security and data integrity.

And sometimes I intentionally break requirements (some of my
domains have no SPF, others have DMARC rua records without
external verification, ...) just to see what happens.

## Commercial Usage {#faq-commercial-usage}

If you want to integrate AboutMy.email with your system via
API, or want to talk about a white-label version, [contact us](/contact).

If you just want to point your customers at our public instance,
feel free.

## Other tools {#faq-tools}

As well as AboutMy.email we have a bunch of other tools that
email folk might find useful.

### [tools.wordtothewise.com](https://tools.wordtothewise.com/)

aka [emailstuff.org](https://emailstuff.org/)

 * General purpose DNS query tools
 * SPF, DKIM and DMARC record checking
 * Base64 and quoted printable decoders
 * Unicode glyph search
 * Pretty RFCs
 * ... all with easily sharable result links you can copy and share

### [reject.wordtothewise.com](https://reject.wordtothewise.com)

This is very useful for testing SMTP implementations and bounce
handling. It lets you send an email and control exactly how it
will be rejected - what rejection message, where in the SMTP
transaction and so on.

### [DKIM Core](https://dkimcore.org)

Help for people deploying DKIM authentication.

### Lots of open source code

I should document this better, but there's a lot of email
related tools on [github/wttw](https://github.com/wttw?tab=repositories)

## Other tools by other people {#faq-other-tools}

 * Al Iverson's [WombatMail](https://www.wombatmail.com)