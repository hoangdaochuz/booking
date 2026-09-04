package kafka

import (
	pkgkafka "github.com/ticketbox/pkg/kafka"
	"github.com/ticketbox/pkg/topics"
	"go.uber.org/zap"
)

type OutboundEventProducer struct {
	logger   *zap.Logger
	producer *pkgkafka.Producer
}

func NewOutboundEventProducer(logger *zap.Logger, brokers []string) *OutboundEventProducer {
	return &OutboundEventProducer{
		logger:   logger,
		producer: pkgkafka.NewProducer(brokers, []string{topics.TopicSchedulerConfigEvent}, logger),
	}
}

// func (op *OutboundEventProducer) PublishSchedulerConfigChanged(ctx context.Context, payload domain.SchedulerConfigChangedPayload) error {

// 	data, err := json.Marshal(payload)
// 	if err != nil {
// 		return fmt.Errorf("Fail to marshal scheduler config event: %w", err)
// 	}
// 	event := pkgkafka.Event{
// 		Type:      topics.TypeSchedulerConfigChanged,
// 		Timestamp: time.Now(),
// 		Data:      data,
// 	}
// 	eventKey := strconv.Itoa(time.Now().Nanosecond()) + uuid.NewString()

// 	return op.producer.Publish(ctx, topics.TopicSchedulerConfigEvent, eventKey, event)
// }

func (o *OutboundEventProducer) Producer() *pkgkafka.Producer {
	return o.producer
}
