package main

import (
	"context"
	cepubsub "github.com/cloudevents/sdk-go/protocol/pubsub/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/kelseyhightower/envconfig"
	"log"
	"os"
)

type envConfig struct {
	ProjectID string `envconfig:"GOOGLE_CLOUD_PROJECT" required:"true"`
	TopicID   string `envconfig:"PUBSUB_TOPIC" default:"demo_cloudevents" required:"true"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		os.Exit(1)
	}

	e := cloudevents.NewEvent()
	e.SetID("42")
	e.SetSubject("max")
	e.SetSource("athena")
	e.SetType("authn")

	//Set custom attribute
	if err := e.Context.SetExtension("transactionid", "dummy_txn_id"); err != nil {
		log.Printf("failed to set custom attribute: %s", err)
		os.Exit(1)
	}

	e.SetData(cloudevents.ApplicationJSON, "foo")

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
