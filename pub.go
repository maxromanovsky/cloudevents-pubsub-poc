package main

import (
	"cloudevents-poc/events"
	"context"
	cepubsub "github.com/cloudevents/sdk-go/protocol/pubsub/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/kelseyhightower/envconfig"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"os"
	"time"
)

type envConfig struct {
	ProjectID string `envconfig:"GOOGLE_CLOUD_PROJECT" required:"true"`
	TopicID   string `envconfig:"PUBSUB_TOPIC" default:"demo_cloudevents" required:"true"`
}

func main() {
	// Read config
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}

	// Set attributes
	e := cloudevents.NewEvent()
	e.SetID("42")
	e.SetSubject("max")
	e.SetSource("athena")
	e.SetType("authn")

	// Set custom attribute
	if err := e.Context.SetExtension("transactionid", "dummy_txn_id"); err != nil {
		log.Printf("failed to set custom attribute: %s", err)
		os.Exit(1)
	}

	// Set protobuf message
	eventData := &events.AuthenticationEvent{
		Context: &events.EventContext{
			Id: "42",
			Attributes: map[string]*events.Value{
				"hello":     {Value: &events.Value_StringValue{StringValue: "world"}},
				"timestamp": {Value: &events.Value_IntegerValue{IntegerValue: time.Now().UTC().Unix()}},
			},
		},
		Result:    events.LoginEventType_LOGIN_EVENT_TYPE_SUCCESS,
		Subject:   "max",
		EventTime: timestamppb.Now(),
	}

	// Marshal protobuf
	eventJson, err := protojson.Marshal(eventData)
	if err != nil {
		log.Printf("failed to marshal protobuf to JSON: %s", err)
		os.Exit(1)
	}

	if err := e.SetData(cloudevents.ApplicationJSON, eventJson); err != nil {
		log.Printf("failed to set event body to JSON: %s", err)
		os.Exit(1)
	}

	// Send event
	t, err := cepubsub.New(context.Background(),
		cepubsub.WithProjectID(env.ProjectID),
		cepubsub.WithTopicID(env.TopicID))
	if err != nil {
		log.Printf("failed to create pubsub transport, %s", err.Error())
		os.Exit(1)
	}
	c, err := cloudevents.NewClient(t, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Printf("failed to create client, %s", err.Error())
		os.Exit(1)
	}

	if result := c.Send(context.Background(), e); cloudevents.IsUndelivered(result) {
		log.Printf("failed to send: %v", err)
		os.Exit(1)
	} else {
		log.Printf("sent, accepted: %t", cloudevents.IsACK(result))
	}

	os.Exit(0)
}
