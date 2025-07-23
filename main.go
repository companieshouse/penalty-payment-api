package main

import (
	"context"
	"errors"
	"fmt"
	gologger "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/common/e5"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/handlers"
	"github.com/companieshouse/penalty-payment-api/penalty_payments/consumer"
	"github.com/gorilla/mux"
	_ "golang.org/x/oauth2"
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
	prDaoService := dao.NewPayableResourcesDaoService(cfg)
	apDaoService := dao.NewAccountPenaltiesDaoService(cfg)

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

	handlers.Register(mainRouter, cfg, prDaoService, apDaoService, penaltyDetailsMap, allowedTransactionsMap)

	if cfg.FeatureFlagPaymentsProcessingEnabled {
		// Push the Sarama logs into our custom writer
		sarama.Logger = gologger.New(&log.Writer{}, "[Sarama] ", gologger.LstdFlags)
		penaltyFinancePayment := &handlers.PenaltyFinancePayment{
			E5Client:                  e5.NewClient(cfg.E5Username, cfg.E5APIURL),
			PayableResourceDaoService: prDaoService,
		}
		go consumer.Consume(cfg, penaltyFinancePayment)
	}

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
			prDaoService.Shutdown()
			os.Exit(1)
		}
	}()

	// wait for app shutdown message before attempting to close server gracefully
	<-stop

	log.Info("shutting down server...")
	prDaoService.Shutdown()
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
