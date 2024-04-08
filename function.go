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

// 建立一個 map 來儲存每個用戶的 ChatSession
var userSessions = make(map[string]*genai.ChatSession)

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

	functions.HTTP("HelloHTTP", helloHTTP)
}

func helloHTTP(w http.ResponseWriter, r *http.Request) {
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
			// Handle only on text message
			case webhook.TextMessageContent:
				req := message.Text
				// 檢查是否已經有這個用戶的 ChatSession or req == "reset"

				// 取得用戶 ID
				var uID string
				switch source := e.Source.(type) {
				case *webhook.UserSource:
					uID = source.UserId
				case *webhook.GroupSource:
					uID = source.UserId
				case *webhook.RoomSource:
					uID = source.UserId
				}

				if _, err := bot.ReplyMessage(
					&messaging_api.ReplyMessageRequest{
						ReplyToken: e.ReplyToken,
						Messages: []messaging_api.MessageInterface{
							&messaging_api.TextMessage{
								Text: fmt.Sprintf("UID: %s - 收到訊息: %s,  ", uID, req),
							},
						},
					},
				); err != nil {
					log.Print(err)
					return
				}

			// Handle only on Sticker message
			case webhook.StickerMessageContent:
				var kw string
				for _, k := range message.Keywords {
					kw = kw + "," + k
				}

				outStickerResult := fmt.Sprintf("收到貼圖訊息: %s, pkg: %s kw: %s  text: %s", message.StickerId, message.PackageId, kw, message.Text)
				if _, err := bot.ReplyMessage(
					&messaging_api.ReplyMessageRequest{
						ReplyToken: e.ReplyToken,
						Messages: []messaging_api.MessageInterface{
							&messaging_api.TextMessage{
								Text: outStickerResult,
							},
						},
					},
				); err != nil {
					log.Print(err)
					return
				}

			// Handle only image message
			case webhook.ImageMessageContent:
				log.Println("Got img msg ID:", message.Id)

				//Get image binary from LINE server based on message ID.
				content, err := blob.GetMessageContent(message.Id)
				if err != nil {
					log.Println("Got GetMessageContent err:", err)
				}
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

				model := client.GenerativeModel("gemini-pro-vision")
				value := float32(0.8)
				model.Temperature = &value
				prompt := []genai.Part{
					genai.ImageData("png", data),
					genai.Text("Describe this image with scientific detail, reply in zh-TW:"),
				}
				log.Println("Begin processing image...")
				resp, err := model.GenerateContent(ctx, prompt...)
				log.Println("Finished processing image...", resp)
				if err != nil {
					log.Fatal(err)
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
