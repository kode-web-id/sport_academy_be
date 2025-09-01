package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ssb_api/config"
	"ssb_api/models"
	"time"
)

func SendFCMNotification(token, title, body string) error {
	serverKey := DotEnv("FCM_SERVER_KEY")

	message := models.FcmMessage{
		To: token,
		Notification: map[string]string{
			"title": title,
			"body":  body,
		},
	}

	jsonData, _ := json.Marshal(message)

	req, err := http.NewRequest("POST", "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "key="+serverKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)
	fmt.Println("Response from FCM:", string(bodyBytes))

	return nil
}

func CreateNotification(userID uint, token, title, body, notifType string) error {
	notif := models.Notification{
		UserID: userID,
		Title:  title,
		Body:   body,
		Type:   notifType,
		Token:  token,
	}

	if err := config.DB.Create(&notif).Error; err != nil {
		return err
	}

	if err := SendFCMNotification(token, title, body); err == nil {
		now := time.Now()
		config.DB.Model(&notif).UpdateColumns(map[string]interface{}{
			"is_sent": true,
			"sent_at": &now,
		})
	} else {
		fmt.Println("Gagal kirim FCM:", err.Error())
	}

	return nil
}
