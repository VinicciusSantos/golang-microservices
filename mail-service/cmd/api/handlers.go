package main

import "net/http"

func (app *Config) sendMail(w http.ResponseWriter, r *http.Request) {
	type mailMessage struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	var (
		requestPayload mailMessage
		err            error
	)

	if err = app.readJSON(w, r, &requestPayload); err != nil {
		app.errorJSON(w, err)
		return
	}

	msg := Message{
		From:    requestPayload.From,
		To:      requestPayload.To,
		Subject: requestPayload.Subject,
		Data:   requestPayload.Message,
	}

	if err = app.Mailer.SendSMTPMessage(msg); err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusOK, jsonResponse{
		Error:  false,
		Message: "Message sent successfully",
	})
}
