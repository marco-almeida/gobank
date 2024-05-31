package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

// UserMessageBrokerRepository represents the repository used for interacting with User records.
type UserMessageBrokerRepository struct {
	client *asynq.Client
}

// NewUserMessageBrokerRepository instantiates the UserMessageBrokerRepository repository.
func NewUserMessageBrokerRepository(redisOpt asynq.RedisClientOpt) *UserMessageBrokerRepository {
	return &UserMessageBrokerRepository{
		client: asynq.NewClient(redisOpt),
	}
}

const TaskSendVerifyEmail = "task:send_verify_email"

func (repo *UserMessageBrokerRepository) CreateVerifyEmailTask(ctx context.Context, username string, opts ...asynq.Option) error {
	jsonPayload, err := json.Marshal(username)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)
	info, err := repo.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("enqueued task")
	return nil
}
