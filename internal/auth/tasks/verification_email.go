package tasks

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const TypeSendVerificationEmail = "email:send_verification"

type SendVerificationEmailPayload struct {
	UserID string `json:"user_id"`
}

func NewSendVerificationEmailTask(userID string) (*asynq.Task, error) {
	payload, err := json.Marshal(SendVerificationEmailPayload{UserID: userID})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSendVerificationEmail, payload), nil
}
