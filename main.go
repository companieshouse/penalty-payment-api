//coverage:ignore file
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
	"github.com/companieshouse/chs.go/kafka/resilience"
	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api/common/dao"
	"github.com/companieshouse/penalty-payment-api/common/e5"
	"github.com/companieshouse/penalty-payment-api/config"
	"github.com/companieshouse/penalty-payment-api/handlers"
	"github.com/companieshouse/penalty-payment-api/penalty_payments/supervisor"
	"github.com/gorilla/mux"
	_ "golang.org/x/oauth2"
)

func main() {
	const exitErrorFormat = "error configuring service: %s. Exiting"
	cfg, err := config.Get()

	if err != nil {
		log.Error(fmt.Errorf(exitErrorFormat, err), nil)
		return
	}

	namespace := cfg.Namespace()
	log.Namespace = namespace
	log.Debug("Config", log.Data{"Config": cfg})

	// Create router
	mainRouter := mux.NewRouter()
	mongoClientProvider, err := dao.NewMongoClient(cfg.MongoDBURL)
	if err != nil {
		log.Error(fmt.Errorf("mongo client error: %s. Exiting", err), nil)
		os.Exit(1)
	}
	prDaoService := dao.NewPayableResourcesDaoService(mongoClientProvider, cfg)
	apDaoService := dao.NewAccountPenaltiesDaoService(mongoClientProvider, cfg)

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
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Push the Sarama logs into our custom writer
		sarama.Logger = gologger.New(&log.Writer{}, "[Sarama] ", gologger.LstdFlags)
		penaltyFinancePayment := &handlers.PenaltyFinancePayment{
			E5Client:                  e5.NewClient(cfg.E5Username, cfg.E5APIURL),
			PayableResourceDaoService: prDaoService,
		}
		go supervisor.SuperviseConsumer(ctx, cfg.ConsumerGroupName, cfg, penaltyFinancePayment, nil)

		retry := &resilience.ServiceRetry{
			ThrottleRate: time.Duration(cfg.ConsumerRetryThrottleRate) * time.Second,
			MaxRetries:   cfg.ConsumerRetryMaxAttempts,
		}
		go supervisor.SuperviseConsumer(ctx, cfg.ConsumerRetryGroupName, cfg, penaltyFinancePayment, retry)
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
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), timeout)
	defer shutdownCancel()

	err = h.Shutdown(shutdownCtx)
	if err != nil {
		log.Error(fmt.Errorf("failed to shutdown server gracefully: [%v]", err))
	} else {
		log.Info("server shutdown gracefully")
	}
}
