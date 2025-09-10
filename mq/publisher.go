package mq

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

func PublishMessage(publisher message.Publisher, topic string, payload []byte) error {
	msg := message.NewMessage(
		watermill.NewUUID(),
		payload,
	)

	msg.Metadata.Set("content-type", "application/json")

	return publisher.Publish(topic, msg)
}