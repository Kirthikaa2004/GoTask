package main

import (
    "fmt"
    "os"
    "flag"

    "room_booking_management/room_booking_service/pkg/server"
    "go.uber.org/zap"
)

func main() {
    env := flag.String("env", "dev", "Environment (dev, qc, uat, prod)")
    flag.Parse()

    app, err := server.InitializeApp(*env)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to initialize app: %v\n", err)
        os.Exit(1)  
    }

    if err := server.RunServer(app); err != nil {
        app.Logger.Fatal("Failed to start the server", zap.Error(err))  
    }
}