package workers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

type SlackOpenSocketResponse struct {
	Ok  bool   `json:"ok"`
	Url string `json:"url"`
}

type SlackEvent struct {
	EnvelopeId             string `json:"envelope_id"`
	Type                   string `json:"type"`
	AcceptsResponsePayload bool   `json:"accepts_response_payload"`
	Payload                struct {
		ApiAppID           string `json:"api_app_id"`
		EventID            string `json:"event_id"`
		EventTime          int    `json:"event_time"`
		Token              string `json:"token"`
		TeamID             string `json:"team_id"`
		Type               string `json:"type"`
		IsExtSharedChannel bool   `json:"is_ext_shared_channel"`
		Event              struct {
			Type    string `json:"type"`
			Channel struct {
				ID             string `json:"id"`
				IsChannel      bool   `json:"is_channel"`
				IsMPIM         bool   `json:"is_mpim"`
				Name           string `json:"name"`
				NameNormalized string `json:"name_normalized"`
				Created        int    `json:"created"`
			} `json:"channel"`
			// TODO: Whats up with this timestamp.
			// EventTS time.Time `json:"event_ts"`
		} `json:"event"`
		Authorizations []struct {
			EnterpriseID        string `json:"enterprise_id"`
			TeamID              string `json:"team_id"`
			UserID              string `json:"user_id"`
			IsBot               bool   `json:"is_bot"`
			IsEnterpriseInstall bool   `json:"is_enterprise_install"`
		} `json:"authorizations"`
	} `json:"payload"`
	RetryAttempt int    `json:"retry_attempt"`
	RetryReason  string `json:"retry_reason"`
}

type SlackEventAcknowledge struct {
	EnvelopeId string `json:"envelope_id"`
}

type SlackSocketMessage struct {
	Type string `json:"type"`
}

type SlackSocketHello struct {
	Type           string `json:"type"`
	NumConnections int    `json:"num_connections"`
	DebugInfo      struct {
		Host                      string `json:"host"`
		BuildNumber               int    `json:"build_number"`
		ApproximateConnectionTime int    `json:"approximate_connection_time"`
	} `json:"debug_info"`
	ConnectionInfo struct {
		AppID string `json:"app_id"`
	} `json:"connection_info"`
}

type SlackSocketDisconnect struct {
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	DebugInfo struct {
		Host string `json:"host"`
	} `json:"debug_info"`
}

type SlackWorker struct {
	appToken   string
	httpClient *http.Client
}

func NewSlackWorker() (*SlackWorker, error) {
	worker := new(SlackWorker)

	appToken, ok := os.LookupEnv("YASMIN_APP_TOKEN")
	if !ok {
		err := errors.New("Missing YASMIN_APP_TOKEN in environment")
		return nil, err
	}
	worker.appToken = appToken

	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}
	worker.httpClient = &http.Client{Transport: tr}

	return worker, nil
}

func (worker *SlackWorker) getWebSocketUrl() (string, error) {
	url := "https://slack.com/api/apps.connections.open"
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+worker.appToken)
	var slackSocketOpenResponse SlackOpenSocketResponse
	resp, err := worker.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	b := []byte(body)
	err = json.Unmarshal(b, &slackSocketOpenResponse)
	if err != nil {
		return "", err
	}
	if !slackSocketOpenResponse.Ok {
		return "", errors.New("error in opening slack websocket")
	}

	return slackSocketOpenResponse.Url, nil
}

func (worker *SlackWorker) DoWork() {
	for {
		wsUrl, err := worker.getWebSocketUrl()
		if err != nil {
			log.Println(err)
			time.Sleep(1 * time.Second) // Try again in 1 second.
			continue
		}
		wsUrl += "&debug_reconnects=true"

		header := make(http.Header)
		conn, _, err := websocket.DefaultDialer.Dial(wsUrl, header)
		if err != nil {
			log.Printf("could not create websocket at this time")
		}
		defer conn.Close()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("failed to read message from the websocket")
				break // Exit the for loop for reading messages, re-establish the connection.
			}

			// Check the message and act on it.
			b := []byte(message)
			debug := string(b[:])
			var slackSocketMessage SlackSocketMessage
			err = json.Unmarshal(b, &slackSocketMessage)
			if err != nil {
				log.Println("could not unmarshal slack socket message", err)
				continue
			}

			// Respond on event type.
			switch slackSocketMessage.Type {
			case "hello":
				var hello SlackSocketHello
				err = json.Unmarshal(b, &hello)
				if err != nil {
					log.Println("hello", err, debug)
				} else {
					log.Printf("%+v\n", hello)
				}
			case "disconnect":
				var disconnect SlackSocketDisconnect
				err = json.Unmarshal(b, &disconnect)
				if err != nil {
					log.Println("disconnect", err, debug)
				} else {
					log.Printf("%+v\n", disconnect)
				}
			case "events_api":
				var event SlackEvent
				err = json.Unmarshal(b, &event)
				if err != nil {
					log.Println("event", err, debug)
				} else {
					log.Printf("%+v\n", event)
					var ack SlackEventAcknowledge
					ack.EnvelopeId = event.EnvelopeId
					conn.WriteJSON(ack)
				}
			default:
				log.Println("unknown", debug)
			}
		}
		time.Sleep(1 * time.Second) // Try again in 1 second if it becomes connected.
	}
}
