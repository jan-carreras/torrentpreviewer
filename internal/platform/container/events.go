package container

import (
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill-googlecloud/pkg/googlecloud"
	"github.com/ThreeDotsLabs/watermill/message"
	"prevtorrent/internal/preview/platform/configuration"
)

type events interface {
	commandPublisher() message.Publisher
	commandSubscriber() message.Subscriber
	eventPublisher() message.Publisher
	eventSubscriber(handlerName string) (message.Subscriber, error)
}

func makeEventDriver(config configuration.Config, log watermill.LoggerAdapter) events {
	switch config.PubSubDriver {
	case "rabbit":
		return newRabbit(config, log)
	case "google":
		return newPubsub(config, log)
	default:
		panic(fmt.Sprintf("unknown PubSubDriver: %v", config.PubSubDriver))
	}
}

type pubsub struct {
	config          configuration.Config
	loggerWatermill watermill.LoggerAdapter
}

func newPubsub(config configuration.Config, loggerWatermill watermill.LoggerAdapter) *pubsub {
	return &pubsub{config: config, loggerWatermill: loggerWatermill}
}

func (p pubsub) commandSubscriber() message.Subscriber {
	googleSubscriber, err := googlecloud.NewSubscriber(
		googlecloud.SubscriberConfig{
			GenerateSubscriptionName: googlecloud.TopicSubscriptionName,
			ProjectID:                p.config.GooglePubSubProjectID,
		},
		p.loggerWatermill,
	)
	if err != nil {
		panic(err)
	}
	return googleSubscriber
}

func (p pubsub) commandPublisher() message.Publisher {
	publisher, err := googlecloud.NewPublisher(googlecloud.PublisherConfig{
		ProjectID: p.config.GooglePubSubProjectID,
	}, p.loggerWatermill)
	if err != nil {
		panic(err)
	}
	return publisher
}

func (p pubsub) eventPublisher() message.Publisher {
	publisher, err := googlecloud.NewPublisher(googlecloud.PublisherConfig{
		ProjectID: p.config.GooglePubSubProjectID,
	}, p.loggerWatermill)
	if err != nil {
		panic(err)
	}
	return publisher
}

func (p pubsub) eventSubscriber(handlerName string) (message.Subscriber, error) {
	return googlecloud.NewSubscriber(
		googlecloud.SubscriberConfig{
			GenerateSubscriptionName: googlecloud.TopicSubscriptionName,
			ProjectID:                p.config.GooglePubSubProjectID,
		},
		p.loggerWatermill,
	)
}

type rabbit struct {
	config          configuration.Config
	loggerWatermill watermill.LoggerAdapter
}

func newRabbit(config configuration.Config, loggerWatermill watermill.LoggerAdapter) *rabbit {
	return &rabbit{config: config, loggerWatermill: loggerWatermill}
}

func (r rabbit) commandSubscriber() message.Subscriber {
	commandsAMQPConfig := amqp.NewDurableQueueConfig(r.config.AMQPURI)
	subs, err := amqp.NewSubscriber(commandsAMQPConfig, r.loggerWatermill)
	if err != nil {
		panic(err)
	}
	return subs
}

func (r rabbit) commandPublisher() message.Publisher {
	commandsAMQPConfig := amqp.NewDurableQueueConfig(r.config.AMQPURI)
	commandsPublisher, err := amqp.NewPublisher(commandsAMQPConfig, r.loggerWatermill)
	if err != nil {
		panic(err)
	}
	return commandsPublisher
}

func (r rabbit) eventPublisher() message.Publisher {
	eventsPublisher, err := amqp.NewPublisher(
		amqp.NewDurablePubSubConfig(r.config.AMQPURI, nil),
		r.loggerWatermill,
	)
	if err != nil {
		panic(err)
	}
	return eventsPublisher
}

func (r rabbit) eventSubscriber(handlerName string) (message.Subscriber, error) {
	config := amqp.NewDurablePubSubConfig(
		r.config.AMQPURI,
		amqp.GenerateQueueNameTopicNameWithSuffix(handlerName),
	)

	return amqp.NewSubscriber(config, r.loggerWatermill)
}
