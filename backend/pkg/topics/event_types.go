package topics

// Event types carried in pkgkafka.Event.Type. Like the topic constants
// above, these are part of the producer/consumer contract — publishers
// set them, consumers switch on them, so both sides must reference
// these constants instead of inlining strings.
const (
	TypeBookingConfirmed = "BookingConfirmed"
	TypeBookingFailed    = "BookingFailed"
	TypeBookingCancelled = "BookingCancelled"

	TypeEventCreated = "EventCreated"

	TypeUserRegistered = "UserRegistered"

	TypePaymentSucceed = "PaymentSucceed"
	TypePaymentFail    = "PaymentFail"

	TypeSchedulerConfigChanged = "SchedulerConfigChanged"
)
