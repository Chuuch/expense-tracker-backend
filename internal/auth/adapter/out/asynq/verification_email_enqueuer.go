package asynq

import (
	"context"

	"github.com/chuuch/expense-tracker-backend/internal/auth/tasks"
	"github.com/hibiken/asynq"
)

type VerificationEmailEnqueuer struct {
	client *asynq.Client
}

func NewVerificationEmailEnqueuer(client *asynq.Client) *VerificationEmailEnqueuer {
	return &VerificationEmailEnqueuer{client: client}
}

func (e *VerificationEmailEnqueuer) EnqueueSendVerificationEmail(ctx context.Context, userID string) error {
	task, err := tasks.NewSendVerificationEmailTask(userID)
	if err != nil {
		return err
	}
	_, err = e.client.EnqueueContext(ctx, task)
	return err
}
