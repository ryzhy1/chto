package main

import (
	"boton-back/internal/app"
	"boton-back/internal/config"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envDev   = "dev"
	envProd  = "prod"
	envLocal = "local"
)

func main() {
	cfg := config.MustLoad()

	fmt.Println(`                                                $$\                               $$\   
                                                $$ |                            $$$$ |  
 $$$$$$\        $$\   $$\       $$$$$$$$\       $$$$$$$\        $$\   $$\       \_$$ |  
$$  __$$\       $$ |  $$ |      \____$$  |      $$  __$$\       $$ |  $$ |        $$ |  
$$ |  \__|      $$ |  $$ |        $$$$ _/       $$ |  $$ |      $$ |  $$ |        $$ |  
$$ |            $$ |  $$ |       $$  _/         $$ |  $$ |      $$ |  $$ |        $$ |  
$$ |            \$$$$$$$ |      $$$$$$$$\       $$ |  $$ |      \$$$$$$$ |      $$$$$$\ 
\__|             \____$$ |      \________|      \__|  \__|       \____$$ |      \______|
                $$\   $$ |                                      $$\   $$ |              
                \$$$$$$  |                                      \$$$$$$  |              
                 \______/                                        \______/               `)

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	log.Info("Starting http", "env", cfg.Server.Env)

	ctx, _ := context.WithCancel(context.Background())

	application := app.New(
		ctx,
		log,
		cfg,
	)

	go application.HTTPServer.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Info("Application stopped", slog.String("signal", sign.String()))

	err := application.HTTPServer.Stop(ctx)
	if err != nil {
		panic(err)
	}
}
