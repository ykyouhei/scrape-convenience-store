package main

import (
	"context"
	"os"

	firebase "firebase.google.com/go"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/ykyouhei/conveniencestore/scraper"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

// MyEvent lambdaからのイベント
type MyEvent struct {
}

// HandleRequest リクエストのハンドラ
func HandleRequest(ctx context.Context, name MyEvent) (string, error) {
	return scrape(ctx)
}

func main() {
	lambda.Start(HandleRequest)
}

func scrape(ctx context.Context) (string, error) {
	credentials, _ := google.CredentialsFromJSON(
		ctx,
		[]byte(os.Getenv("FIREBASE_CREDENTIALS")),
		"https://www.googleapis.com/auth/cloud-platform",
		"https://www.googleapis.com/auth/datastore",
		"https://www.googleapis.com/auth/devstorage.full_control",
		"https://www.googleapis.com/auth/firebase",
		"https://www.googleapis.com/auth/identitytoolkit",
		"https://www.googleapis.com/auth/userinfo.email")

	opt := option.WithCredentials(credentials)
	conf := &firebase.Config{DatabaseURL: os.Getenv("FIREBASE_DATABASE_URL")}
	app, _ := firebase.NewApp(ctx, conf, opt)

	client, _ := app.Database(ctx)

	scrapers := []scraper.Scraper{
		scraper.LawsonScraper{},
		scraper.FamilyMartScraper{},
		scraper.SevenElevenScraper{}}

	for _, scraper := range scrapers {
		ref := client.NewRef(scraper.DatabasePath())
		items := scraper.Scrape()

		if error := ref.Set(ctx, items); error != nil {
			return "database set error", error
		}
	}

	return "success", nil
}
