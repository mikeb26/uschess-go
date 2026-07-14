# uschess-go

[![Go Reference](https://pkg.go.dev/badge/github.com/mikeb26/uschess-go.svg)](https://pkg.go.dev/github.com/mikeb26/uschess-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

An unofficial, community-maintained Go client for the [US Chess Ratings API](https://ratings-api.uschess.org/swagger/v1/swagger.json).

`uschess-go` provides generated, response-aware bindings for the API, plus practical conveniences for JSON requests, retrying eligible read requests, and collecting paginated results.

> **Unofficial project:** This library is not affiliated with, endorsed by, or supported by US Chess. The API and its availability, behavior, and schema are controlled by US Chess.

## Features

- Typed Go client generated from the US Chess Ratings API OpenAPI specification
- Response-aware methods that expose HTTP status, headers, raw bodies, and decoded payloads
- Default client configured for JSON responses
- Automatic retries for eligible read-only requests on transient failures and retryable HTTP statuses
- Context-aware requests and retry delays
- Helpers that retrieve all pages for common member, affiliate, event, section, standings, and Grand Prix queries
- Custom request editors and HTTP client support through the generated client options

## Requirements

- Go **1.26.5** or later

## Installation

```sh
go get github.com/mikeb26/uschess-go
```

## Quick start

Look up a member by US Chess ID:

```go
package main

import (
    "context"
    "fmt"
    "log"

    uschess "github.com/mikeb26/uschess-go"
)

func main() {
    ctx := context.Background()

    client, err := uschess.NewDefaultClient()
    if err != nil {
        log.Fatal(err)
    }

    response, err := client.GetMemberWithResponse(ctx, "12641216")
    if err != nil {
        log.Fatal(err)
    }
    if response.JSON200 == nil {
        log.Fatalf("unexpected response: HTTP %d: %s", response.StatusCode(), response.Body)
    }

    member := response.JSON200
    fmt.Printf("%s %s (ID: %s)\n", member.FirstName, member.LastName, member.Id)

    for _, rating := range member.Ratings {
        if rating.RatingType == uschess.RatingTypeR {
            fmt.Printf("Regular rating: %d\n", rating.Rating)
            break
        }
    }
}
```

## Examples

Runnable examples are available in the [`examples`](examples) directory, covering common member, event, and affiliate lookups. For example:

```sh
go run ./examples/member
```

See [`examples`](examples) for the complete list and source.

## Usage

### Default client

`NewDefaultClient` targets `https://ratings-api.uschess.org`, requests `application/json` responses by default, and adds retry behavior to eligible idempotent read requests. Retryable responses include `408`, `425`, `429`, `502`, `503`, and `504`; delays respect `Retry-After` when present.

All client calls accept a `context.Context`. Use a deadline or cancellation signal to bound both requests and retry waits:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

response, err := client.GetAffiliateWithResponse(ctx, uschess.AffiliateID("A1234567"))
```

### Response-aware API methods

The generated `*WithResponse` methods return both transport details and a decoded success payload. Always check the expected decoded response before using it:

```go
response, err := client.GetRatedEventWithResponse(ctx, uschess.EventID("202501010001"))
if err != nil {
    return err
}
if response.JSON200 == nil {
    return fmt.Errorf("ratings API returned HTTP %d: %s", response.StatusCode(), response.Body)
}

event := response.JSON200
```

See the [Go package documentation](https://pkg.go.dev/github.com/mikeb26/uschess-go) for every generated operation and model.

### Retrieve all pages

Convenience methods follow pagination until the API indicates there is no next page. For example:

```go
members, err := client.GetAllMembers(ctx)
if err != nil {
    return err
}

sections, err := client.GetAllRatedSections(ctx)
if err != nil {
    return err
}
```

Available helpers include:

- `GetAllAffiliates`
- `GetAllMembers`, `GetAllMemberAwards`, `GetAllMemberDirectorships`, `GetAllMemberRatedEvents`, `GetAllMemberRatedGames`, `GetAllMemberRatedSections`, `GetAllRatingSupplements`, and `GetAllTopPlayersReportsForMember`
- `GetAllRatedEvents`, `GetAllRatedSections`, and `GetAllRatedEventStandings`
- `GetAllAffiliateRatedEvents`
- `GetAllGrandPrixStandings` and `GetAllGrandPrixSections`
- `GetAllPendingEvents` and `GetAllPendingPlayers`

These methods may issue many requests and return large result sets. Prefer the generated page methods when you need a bounded result set or direct control over offsets. Helpers that accept parameter structs, including `GetAllMemberRatedGames`, also support the endpoint's filters and page-size option.

### Custom HTTP behavior

The generated client accepts [`ClientOption`](https://pkg.go.dev/github.com/mikeb26/uschess-go#ClientOption) values. For example, provide an HTTP client with your own timeout:

```go
httpClient := &http.Client{Timeout: 15 * time.Second}
client, err := uschess.NewDefaultClient(uschess.WithHTTPClient(httpClient))
if err != nil {
    return err
}
```

You can also pass request editors to add headers or otherwise modify individual requests. Refer to the generated client documentation for the available options and method signatures.

## API coverage

The generated client is derived from the API specification committed as [`swagger.json`](swagger.json). It includes the operations and schemas described there, while this package adds the default client configuration and pagination helpers.

To regenerate the client after updating the specification:

```sh
make build
```

The project uses [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) for generation. Avoid hand-editing [`client.gen.go`](client.gen.go); update the OpenAPI specification or generation configuration instead.

## Development

```sh
# Run tests
go test ./...

# Build the package and examples
make build
```

## Contributing

Contributions are welcome. Please open an issue to discuss substantial changes, keep changes focused, add or update tests where appropriate, and run `go test ./...` before submitting a pull request.

When reporting an API issue, include the endpoint, response status, and a minimal reproducible example—but do not include credentials or other sensitive information.

## License

Distributed under the [MIT License](LICENSE).
