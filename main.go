package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/dao"
	"github.com/companieshouse/penalty-payment-api/handlers"
	"github.com/gorilla/mux"
)

func main() {
	namespace := "penalty-payment-api"
	log.Namespace = namespace

	const exitErrorFormat = "error configuring service: %s. Exiting"
	cfg, err := config.Get()

	if err != nil {
		log.Error(fmt.Errorf(exitErrorFormat, err), nil)
		return
	}

	// Create router
	mainRouter := mux.NewRouter()
	svc := dao.NewDAOService(cfg)

	penaltyDetailsMap, err := config.LoadPenaltyDetails("assets/penalty_details.yml")
	if err != nil {
		log.Error(fmt.Errorf(exitErrorFormat, err), nil)
		return
	}

	allowedTransactionsMap, err := config.GetAllowedTransactions("assets/penalty_types.yml")
	if err != nil {
		log.Error(fmt.Errorf(exitErrorFormat, err), nil)
		return
	}

	handlers.Register(mainRouter, cfg, svc, penaltyDetailsMap, allowedTransactionsMap)

	log.Info("Starting " + namespace)

	h := &http.Server{
		Addr:    cfg.BindAddr,
		Handler: mainRouter,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// run server in new go routine to allow app shutdown signal wait below
	go func() {
		log.Info("starting server...", log.Data{"port": cfg.BindAddr})
		err = h.ListenAndServe()
		log.Info("server stopping...")
		if err != nil && !errors.Is(http.ErrServerClosed, err) {
			log.Error(err)
			svc.Shutdown()
			os.Exit(1)
		}
	}()

	// wait for app shutdown message before attempting to close server gracefully
	<-stop

	log.Info("shutting down server...")
	svc.Shutdown()
	timeout := time.Duration(5) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = h.Shutdown(ctx)
	if err != nil {
		log.Error(fmt.Errorf("failed to shutdown server gracefully: [%v]", err))
	} else {
		log.Info("server shutdown gracefully")
	}
}
