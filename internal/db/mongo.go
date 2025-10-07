package db

import (
	"context"
	"errors"
	"time"

	"github.com/frangi01/bbtelgo/internal/config"
	"github.com/frangi01/bbtelgo/internal/logx"
	"github.com/frangi01/bbtelgo/internal/repo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewClient(ctx context.Context, config config.MongoCfg) (*mongo.Client, error) {
	if config.URI == "" {
		return nil, errors.New("empty Mongo URI")
	}

	// client options
	opts := options.Client().ApplyURI(config.URI)

	if config.AppName != "" {
		opts.SetAppName(config.AppName)
	}
	if config.MinPoolSize > 0 {
		opts.SetMinPoolSize(config.MinPoolSize)
	}
	if config.MaxPoolSize > 0 {
		opts.SetMaxPoolSize(config.MaxPoolSize)
	}
	if config.MaxConnectingLimit > 0 {
		opts.SetMaxConnecting(config.MaxConnectingLimit)
	}
	 

	// Timeout connection (handshake + server selection)
	if config.ConnectTimeout > 0 {
		opts.SetConnectTimeout(config.ConnectTimeout)
		opts.SetServerSelectionTimeout(config.ConnectTimeout)
	}

	// sane defaults
	opts.SetRetryReads(true)
	opts.SetRetryWrites(true)
	
	// optionals:
	// opts.SetReadPreference(readpref.Primary())
	// opts.SetWriteConcern(writeconcern.Majority())

	// connect with timeout
	dialCtx := ctx
	var cancelDial context.CancelFunc
	if config.ConnectTimeout > 0 {
		dialCtx, cancelDial = context.WithTimeout(ctx, config.ConnectTimeout)
		defer cancelDial()
	}

	client, err := mongo.Connect(dialCtx, opts)
	if err != nil {
		return nil, err
	}

	// ping with timeout "command" (CmdTimeout)
	pingTimeout := config.CmdTimeout
	if pingTimeout <= 0 {
		pingTimeout = 10 * time.Second
	}
	pingCtx, cancelPing := context.WithTimeout(ctx, pingTimeout)
	defer cancelPing()

	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}

	return client, nil
}

type RepositoryList struct {
	UserRepository 		*repo.UserRepository
	MessageRepository 	*repo.MessageRepository
}

func NewRepositoryList(config config.MongoCfg, client *mongo.Client, logger *logx.Logger) (*RepositoryList, error) {
	userrepo, err := repo.NewUserRepository(client, config.DB)
	if err != nil {
		logger.Errorf("repo init: %v", err)
	}

	msgrepo, err := repo.NewMessageRepository(client, config.DB)
	if err != nil {
		logger.Errorf("repo init: %v", err)
	}

	return &RepositoryList{UserRepository: userrepo, MessageRepository: msgrepo}, err
}
