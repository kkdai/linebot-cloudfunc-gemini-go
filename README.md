# LINE Bot Server in Go

This repository contains a LINE Bot server written in Go. The server is designed to respond to various types of LINE messages including text, sticker, and image messages.

## Functionality

The server is initialized in the `init` function, where it reads environment variables for the Google Gemini API key and the LINE channel access token. It then creates instances of the LINE Messaging API and the LINE Messaging Blob API.

The server's main function is `HelloHTTP`, which is registered as an HTTP Cloud Function. This function parses incoming requests from LINE, verifies the signature, and handles the events in the request.

The server can handle the following types of events:

- Text messages: The server responds with the user ID and the received message.
- Sticker messages: The server responds with the details of the received sticker.
- Image messages: The server retrieves the image from the LINE server, sends it to the Google Gemini API for processing, and responds with the result.
- Follow events: The server logs that it has been followed.
- Postback events: The server logs the postback data.
- Beacon events: The server logs the beacon hardware ID.

## Dependencies

This server uses the `functions-framework-go` package from Google Cloud Platform to register the function, the `line-bot-sdk-go` package to interact with the LINE Messaging API, and the `generative-ai-go` package to interact with the Google Gemini API. It also uses the standard `net/http`, `fmt`, `log`, `os`, `io`, and `context` packages from the Go standard library.

## Usage

To use this server, set the `GOOGLE_GEMINI_API_KEY` and `ChannelAccessToken` environment variables to your Google Gemini API key and LINE channel access token, respectively. Then, send LINE messages to the bot associated with the channel access token. The server will respond according to the type of message.
