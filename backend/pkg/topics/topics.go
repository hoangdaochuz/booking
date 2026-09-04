package topics

// Kafka topics used across services. Producers and consumers must
// reference these constants instead of inlining topic strings so the
// publisher/subscriber contract lives in one place.
const (
	TopicBookingEvents        = "booking.events"
	TopicEventEvents          = "event.events"
	TopicUserEvents           = "user.events"
	TopicPaymentEvents        = "payment.events"
	TopicOrderEvents          = "order.events"
	TopicNotificationEvents   = "notification.events"
	TopicSchedulerConfigEvent = "scheduler.config.events"
)
