package main

import (
	"context"
	"fmt"
)

// MyEvent イベント
type MyEvent struct {
	Name string `json:"name"`
}

// HandleRequest リクエストのハンドラ
func HandleRequest(ctx context.Context, name MyEvent) (string, error) {
	return fmt.Sprintf("Hello %s!", name.Name), nil
}

func main() {
	// lambda.Start(HandleRequest)
	ScrapeToSevenEleven()

}
