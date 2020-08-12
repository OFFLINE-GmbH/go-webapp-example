package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-webapp-example/internal/app"
	"go-webapp-example/pkg/log"

	"github.com/spf13/cobra"
)

var start time.Time

// nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(serveCmd)
	start = time.Now()
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Server",
	Long:  `This command boots the web server and serves the application to the local network.`,
	Run:   runServer,
}

func runServer(_ *cobra.Command, _ []string) {
	instance := boot()

	server := &http.Server{
		Addr:         instance.Config.Server.URL(),
		Handler:      instance,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger := instance.Log.WithPrefix("cmd.serve")
	logger.Debugf("environment is %s", instance.Config.App.Environment)

	err := instance.Migrate()
	if err != nil {
		logger.Fatalf(fmt.Sprintf("failed to run migrations: %s", err))
		return
	}

	go instance.Listen(server)
	go instance.StartDaemons()

	logger.WithFields(log.Fields{"startup.time": time.Since(start)}).Debugf("app started")

	graceful(instance, 30*time.Second, logger)
}

func graceful(instance *app.Kernel, timeout time.Duration, logger log.Logger) {
	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := instance.Shutdown(ctx); err != nil {
		logger.Errorf("application shutdown error: %v\n", err)
	} else {
		logger.Infoln("application stopped")
	}
}
