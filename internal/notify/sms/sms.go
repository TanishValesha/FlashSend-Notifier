package sms

import (
	"fmt"

	"github.com/TanishValesha/FlashSend-Notifier/internal/config"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	"github.com/vonage/vonage-go-sdk"
)

type SMSRequest struct {
	To   string `json:"to"`
	Body string `json:"body"`
}

func SendSMSWithFailover(to string, body string) error {
	err := SendSMS(to, body)
	if err == nil {
		return nil
	}

	fmt.Println("Twilio failed, trying Vonage... Error:", err)

	err = SendViaVonage(to, body)
	if err == nil {
		return nil
	}

	return fmt.Errorf("both providers failed: %v", err)
}

func SendSMS(to string, body string) error {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: config.Cfg.TwilioAccountSID,
		Password: config.Cfg.TwilioAuthToken,
	})

	params := &openapi.CreateMessageParams{}
	params.SetTo(to)
	params.SetFrom(config.Cfg.TwilioPhoneNumber)
	params.SetBody(body)

	_, err := client.Api.CreateMessage(params)
	return err
}

func SendViaVonage(to string, body string) error {
	auth := vonage.CreateAuthFromKeySecret(
		config.Cfg.VonageAPIKey,
		config.Cfg.VonageAPISecret,
	)

	client := vonage.NewSMSClient(auth)

	response, _, err := client.Send(
		config.Cfg.VonageFrom,
		to,
		body,
		vonage.SMSOpts{},
	)

	if err != nil {
		return err
	}

	if response.Messages[0].Status != "0" {
		return fmt.Errorf("vonage failed with status %s", response.Messages[0].Status)
	}

	return nil
}
