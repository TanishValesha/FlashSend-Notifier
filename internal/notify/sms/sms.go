package sms

import (
	"github.com/TanishValesha/FlashSend-Notifier/internal/config"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type SMSRequest struct {
	To   string `json:"to"`
	Body string `json:"body"`
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
