package helloworld

import (
	"context"
	"fmt"
	"net/http"

	"io"
	"log"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"google.golang.org/api/option"

	"github.com/google/generative-ai-go/genai"
	"github.com/line/line-bot-sdk-go/v8/linebot"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

var bot *messaging_api.MessagingApiAPI
var blob *messaging_api.MessagingApiBlobAPI
var geminiKey string
var channelToken string

func init() {
	var err error
	geminiKey = os.Getenv("GOOGLE_GEMINI_API_KEY")
	channelToken = os.Getenv("ChannelAccessToken")
	bot, err = messaging_api.NewMessagingApiAPI(channelToken)
	if err != nil {
		log.Fatal(err)
	}

	blob, err = messaging_api.NewMessagingApiBlobAPI(channelToken)
	if err != nil {
		log.Fatal(err)
	}

	functions.HTTP("HelloHTTP", HelloHTTP)
}

func HelloHTTP(w http.ResponseWriter, r *http.Request) {
	cb, err := webhook.ParseRequest(os.Getenv("ChannelSecret"), r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range cb.Events {
		log.Printf("Got event %v", event)
		switch e := event.(type) {
		case webhook.MessageEvent:
			switch message := e.Message.(type) {

			// Handle only text messages
			case webhook.TextMessageContent:
				req := message.Text

				ctx := context.Background()
				client, err := genai.NewClient(ctx, option.WithAPIKey(geminiKey))
				if err != nil {
					log.Fatal(err)
				}
				defer client.Close()

				// Pass the text content to the gemini-pro model for text generation
				model := client.GenerativeModel("gemini-pro")
				resp, err := model.GenerateContent(ctx, genai.Text(req))
				if err != nil {
					log.Fatal(err)
				}
				var ret string
				for _, cand := range resp.Candidates {
					for _, part := range cand.Content.Parts {
						ret = ret + fmt.Sprintf("%v", part)
						log.Println(part)
					}
				}

				if _, err := bot.ReplyMessage(
					&messaging_api.ReplyMessageRequest{
						ReplyToken: e.ReplyToken,
						Messages: []messaging_api.MessageInterface{
							&messaging_api.TextMessage{
								Text: ret,
							},
						},
					},
				); err != nil {
					log.Print(err)
					return
				}

			// Handle only image messages
			case webhook.ImageMessageContent:
				log.Println("Got img msg ID:", message.Id)

				// Get image content through message.Id
				content, err := blob.GetMessageContent(message.Id)
				if err != nil {
					log.Println("Got GetMessageContent err:", err)
				}
				// Read image content
				defer content.Body.Close()
				data, err := io.ReadAll(content.Body)
				if err != nil {
					log.Fatal(err)
				}
				ctx := context.Background()
				client, err := genai.NewClient(ctx, option.WithAPIKey(geminiKey))
				if err != nil {
					log.Fatal(err)
				}
				defer client.Close()

				// Pass the image content to the gemini-pro-vision model for image description
				model := client.GenerativeModel("gemini-pro-vision")
				prompt := []genai.Part{
					genai.ImageData("png", data),
					genai.Text("Describe this image with scientific detail, reply in zh-TW:"),
				}
				resp, err := model.GenerateContent(ctx, prompt...)
				if err != nil {
					log.Fatal(err)
				}

				// Get the returned content
				var ret string
				for _, cand := range resp.Candidates {
					for _, part := range cand.Content.Parts {
						ret = ret + fmt.Sprintf("%v", part)
						log.Println(part)
					}
				}

				// Reply message
				if _, err := bot.ReplyMessage(
					&messaging_api.ReplyMessageRequest{
						ReplyToken: e.ReplyToken,
						Messages: []messaging_api.MessageInterface{
							&messaging_api.TextMessage{
								Text: ret,
							},
						},
					},
				); err != nil {
					log.Print(err)
					return
				}

			// Handle only video message
			case webhook.VideoMessageContent:
				log.Println("Got video msg ID:", message.Id)

			default:
				log.Printf("Unknown message: %v", message)
			}
		case webhook.FollowEvent:
			log.Printf("message: Got followed event")
		case webhook.PostbackEvent:
			data := e.Postback.Data
			log.Printf("Unknown message: Got postback: " + data)
		case webhook.BeaconEvent:
			log.Printf("Got beacon: " + e.Beacon.Hwid)
		}
	}
}
