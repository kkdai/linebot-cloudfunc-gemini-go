# Hello World HTTP Cloud Function in Go

This repository contains a simple HTTP Cloud Function written in Go. The function is designed to respond to HTTP requests with a personalized greeting.

## Functionality

The function `helloHTTP` is an HTTP Cloud Function that takes a request parameter. It attempts to decode the JSON body of the request into a struct with a single field, `Name`. If the decoding is successful and the `Name` field is not empty, it responds with a personalized greeting: "Hello, [Name]!". If the decoding fails or the `Name` field is empty, it responds with "Hello, World!".

The function is registered in the `init` function, which is automatically called when the package is initialized. The function is registered with the name "HelloHTTP".

## Dependencies

This function uses the `functions-framework-go` package from Google Cloud Platform to register the function, and the standard `net/http`, `fmt`, `html`, and `encoding/json` packages from the Go standard library.

## Usage

To use this function, send an HTTP request with a JSON body containing a `name` field. For example:

```json
{
    "name": "Alice"
}
