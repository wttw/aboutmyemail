# aboutmyemail

API, utilities and content for [AboutMy.email](https://aboutmy.email/).

## API

You need an API key to use the API. There's currently no automated way for you to get an API key, so unless you're
already discussing this with the developers then the API, and the utilities that use it, aren't going to be very useful
to you.

The API for submitting emails for processing (e.g. for ad-hoc testing of content or ESP pre-flight tools), and
for customizing whitelabel versions of the app is described in ameapi.yml in OpenAPI format, suitable for mechanical
generation of client code.

As a Go module github.com/wttw/aboutmyemail provides a Go client implementation of the API. See api.go for the
developer-friendly entrypoints.

## Utilities

Two small commandline utilities are included that use the API. `aboutmyemail` will submit a message for processing,
`setbranding` allows uploading content to a white label site.

Both utilities require an API key to use.

Binary builds of both should be available under the Releases link.

## Content

The text content of the site is implemented in markdown files in ./content. If you're developing white label branding
this is where to start with your content.

## Localization

In ./locale are the translations into the languages we support, and the list of strings to be translated in
messages.pot. This is all standard gettext style files; if you're doing translation and don't know where to
start I'd suggest [Poedit](https://poedit.net).

## Issues

This is also where we manage issues for the AboutMy.email codebase, feel free to
[report bugs or ask for enhancements](https://github.com/wttw/aboutmyemail/issues).
