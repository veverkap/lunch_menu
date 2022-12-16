package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type AutoGenerated struct {
	LUNCH []LUNCH `json:"LUNCH"`
}
type LUNCH struct {
	MenuItemDescription string `json:"MenuItemDescription"`
}

// Message represents a Telegram message.
type Message struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

func main() {
	servingDate := time.Now().AddDate(0, 0, 1)
	log.Printf("Loading menu for %s", servingDate)

	str := "https://webapis.schoolcafe.com/api/CalendarView/GetDailyMenuitems?SchoolId=ccff3367-7f5f-4a0d-a8cf-89e1afafe4ba&ServingDate="
	str = str + servingDate.Format("01/02/2006")

	str = str + "&ServingLine=Standard%20Line&MealType=Lunch"
	fmt.Println(str)

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, str, nil)
	if err != nil {
		log.Fatal(err)
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	g := AutoGenerated{}
	jsonErr := json.Unmarshal(body, &g)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	lunch := fmt.Sprintf("Lunch for ")
	for _, v := range g.LUNCH {
		lunch = lunch + "\r\n" + v.MenuItemDescription
	}

	fmt.Println(lunch)
	botToken := "1388326080:AAFGxulzcVRIJwSCcQr1pGjddyOwvC5_Fe0"
	telegram_url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	fmt.Println(telegram_url)

	// // Create a new message.
	// message := &Message{
	// 	ChatID: -1001675706309,
	// 	Text:   lunch,
	// }
	// SendMessage(telegram_url, message)
}

// SendMessage sends a message to given URL.
func SendMessage(url string, message *Message) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	response, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	defer func(body io.ReadCloser) {
		if err := body.Close(); err != nil {
			log.Println("failed to close response body")
		}
	}(response.Body)
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send successful request. Status was %q", response.Status)
	}
	return nil
}
