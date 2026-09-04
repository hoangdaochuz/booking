package cronjob

import "context"

type ReservationCleanerJob struct {
}

func NewReservationCleanerJob() *ReservationCleanerJob {
	return &ReservationCleanerJob{}
}

func (r *ReservationCleanerJob) Name() string {
	return "reservation-cleaner-job"
}

func (r *ReservationCleanerJob) Run(ctx context.Context) error {
	return nil
}
