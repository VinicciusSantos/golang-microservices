package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/rpc"
	"time"

	"github.com/VinicciusSantos/golang-microservices/broker-service/cmd/api/event"
	"github.com/VinicciusSantos/golang-microservices/broker-service/logs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type requestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	})
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload requestPayload

	if err := app.readJSON(w, r, &requestPayload); err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	case "log":
		app.logItemViaRPC(w, requestPayload.Log)
	case "mail":
		app.sendMail(w, requestPayload.Mail)
	default:
		app.errorJSON(w, errors.New("unknown action"), http.StatusBadRequest)
	}
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	var (
		jsonData, _  = json.MarshalIndent(a, "", "\t")
		client       = &http.Client{}
		request      *http.Request
		response     *http.Response
		authResponse jsonResponse
		err          error
	)

	if request, err = http.NewRequest(http.MethodPost, "http://authentication-service:9091/authenticate", bytes.NewBuffer(jsonData)); err != nil {
		app.errorJSON(w, err)
		return
	}

	if response, err = client.Do(request); err != nil {
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("authentication service error"), http.StatusInternalServerError)
		return
	}

	if err := json.NewDecoder(response.Body).Decode(&authResponse); err != nil {
		app.errorJSON(w, err)
		return
	}

	if authResponse.Error {
		app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	app.writeJSON(w, http.StatusOK, jsonResponse{
		Error:   false,
		Message: "Authenticated",
		Data:    authResponse.Data,
	})
}

func (app *Config) logItem(w http.ResponseWriter, l LogPayload) {
	var (
		jsonData, _ = json.MarshalIndent(l, "", "\t")
		client      = &http.Client{}
		request     *http.Request
		response    *http.Response
		err         error
	)

	if request, err = http.NewRequest(http.MethodPost, "http://logger-service:9092/log", bytes.NewBuffer(jsonData)); err != nil {
		app.errorJSON(w, err)
		return
	}

	if response, err = client.Do(request); err != nil {
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("logging service error"), http.StatusInternalServerError)
		return
	}

	app.writeJSON(w, http.StatusAccepted, jsonResponse{
		Error:   false,
		Message: "logged",
	})
}

func (app *Config) sendMail(w http.ResponseWriter, m MailPayload) {
	var (
		jsonData, _ = json.MarshalIndent(m, "", "\t")
		client      = &http.Client{}
		request     *http.Request
		response    *http.Response
		err         error
	)

	if request, err = http.NewRequest(http.MethodPost, "http://mail-service:9093/send", bytes.NewBuffer(jsonData)); err != nil {
		app.errorJSON(w, err)
		return
	}

	if response, err = client.Do(request); err != nil {
		app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJSON(w, errors.New("mail service error"), http.StatusInternalServerError)
		return
	}

	app.writeJSON(w, http.StatusAccepted, jsonResponse{
		Error:   false,
		Message: "mail sent",
	})
}

func (app *Config) logEventViaRabbit(w http.ResponseWriter, l LogPayload) {
	if err := app.pushToQueue(l.Name, l.Data); err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusAccepted, jsonResponse{
		Error:   false,
		Message: "logged via rabbit",
	})
}

func (app *Config) pushToQueue(name, msg string) (err error) {
	var (
		emmiter     event.Emitter
		jsonData, _ = json.MarshalIndent(LogPayload{
			Name: name,
			Data: msg,
		}, "", "\t")
	)

	if emmiter, err = event.NewEventEmitter(app.Rabbit); err != nil {
		return err
	}

	return emmiter.Push(string(jsonData), "log.INFO")
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) logItemViaRPC(w http.ResponseWriter, l LogPayload) {
	var (
		client     *rpc.Client
		err        error
		response   string
		rpcPayload = RPCPayload(l)
	)

	if client, err = rpc.Dial("tcp", "logger-service:5001"); err != nil {
		app.errorJSON(w, err)
		return
	}

	if err = client.Call("RPCServer.LogInfo", rpcPayload, &response); err != nil {
		app.errorJSON(w, err)
		return
	}

	app.writeJSON(w, http.StatusAccepted, jsonResponse{
		Error:   false,
		Message: response,
	})
}

func (app *Config) LogViaGRPC(w http.ResponseWriter, r *http.Request) {
	var (
		ctx, cancel    = context.WithTimeout(context.Background(), time.Second)
		requestPayload LogPayload
		conn           *grpc.ClientConn
		err            error
	)

	defer cancel()

	if err = app.readJSON(w, r, &requestPayload); err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	if conn, err = grpc.NewClient("logger-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials())); err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	client := logs.NewLoggerServiceClient(conn)

	if _, err = client.WriteLog(ctx, &logs.LogRequest{
		LogEntry: &logs.Log{
			Name: requestPayload.Name,
			Data: requestPayload.Data,
		},
	}); err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	app.writeJSON(w, http.StatusAccepted, jsonResponse{
		Error:   false,
		Message: "logged via gRPC",
	})
}
