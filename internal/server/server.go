package server

import (
	"fmt"
	"net"
	"net/http"

	"github.com/elug3/gochat/internal/config"
	"github.com/elug3/gochat/internal/handler"
	"github.com/elug3/gochat/pkg/service"
	cstore "github.com/elug3/gochat/pkg/store/contacts/sqlite"
	ustore "github.com/elug3/gochat/pkg/store/user/sqlite"
)

func SetupServer(cfg *config.Config) (*http.Server, error) {
	addr := net.JoinHostPort("localhost", fmt.Sprintf("%d", cfg.Port))
	// saveDir := cfg.SaveDir

	// store
	userStore, err := ustore.NewUserStore(cfg)
	if err != nil {
		return nil, err
	}
	contactsStore, err := cstore.NewContactsStore(cfg)
	if err != nil {
		return nil, err
	}
	// event

	// service
	userService, err := service.NewUserService(userStore)
	if err != nil {
		return nil, fmt.Errorf("NewUserService: %w", err)
	}
	contactsService, err := service.NewContactsService(contactsStore)
	if err != nil {
		return nil, fmt.Errorf("NewContactsService: %w", err)
	}

	userHandler, err := handler.NewUserHandler(userService)
	if err != nil {
		return nil, fmt.Errorf("NewUserHandler: %w", err)
	}
	authHandler, err := handler.NewAuthHandler(userService)
	if err != nil {
		return nil, fmt.Errorf("NewAuthHandler: %w", err)
	}

	groupHandler, err := handler.NewGroupHandler(contactsService)
	if err != nil {
		return nil, fmt.Errorf("NewContactsHandler: %w", err)
	}
	r := handler.SetupRoutes(
		userHandler,
		authHandler,
		groupHandler,
	)
	{
		// testing
		contactsService.CreateProfile(1, "test")
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	return srv, nil
}
