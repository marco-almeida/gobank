package redis

import (
	"context"

	"github.com/hibiken/asynq"
	redisRepo "github.com/marco-almeida/mybank/internal/redis"
	"github.com/marco-almeida/mybank/internal/service"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

type TaskProcessor interface {
	Start() error
	Shutdown()
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server          *asynq.Server
	emailService    service.EmailService
	userRepo        service.UserRepository
	verifyEmailRepo service.VerifyEmailRepository
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, emailService service.EmailService, userRepo service.UserRepository, verifyEmailRepo service.VerifyEmailRepository) TaskProcessor {
	logger := NewLogger()
	redis.SetLogger(logger)

	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().Err(err).Str("type", task.Type()).
					Bytes("payload", task.Payload()).Msg("process task failed")
			}),
			Logger: logger,
		},
	)

	return &RedisTaskProcessor{
		server:          server,
		emailService:    emailService,
		userRepo:        userRepo,
		verifyEmailRepo: verifyEmailRepo,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	// register tasks handlers
	mux.HandleFunc(redisRepo.TaskSendVerifyEmail, processor.ProcessTaskSendVerifyEmail)

	return processor.server.Start(mux)
}

func (processor *RedisTaskProcessor) Shutdown() {
	processor.server.Shutdown()
}
