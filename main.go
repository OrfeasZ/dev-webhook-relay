package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	serverCmd := flag.NewFlagSet("server", flag.ExitOnError)
	serverPort := serverCmd.Uint("port", 8080, "The port to run the server on")
	serverHost := serverCmd.String("host", "0.0.0.0", "The host to bind the server to")
	serverHttpEndpoint := serverCmd.String("http-endpoint", "/webhook", "The HTTP endpoint to listen for incoming webhooks on")
	serverToken := serverCmd.String("token", "", "The authentication token used to authorize websocket connections")

	clientCmd := flag.NewFlagSet("client", flag.ExitOnError)
	clientServer := clientCmd.String("server", "", "The server to connect to")
	clientToken := clientCmd.String("token", "", "The authentication token used to authenticate with the server")
	clientForwardUrl := clientCmd.String("forward-url", "", "The URL to forward requests to")

	if len(os.Args) < 2 {
		println("Starting a webhook relay server: dev-webhook-relay server [options]\nOptions:")
		serverCmd.PrintDefaults()
		println("\nStarting a webhook relay client: dev-webhook-relay client [options]\nOptions:")
		clientCmd.PrintDefaults()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		err := serverCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}

		if *serverPort < 1 || *serverPort > 65535 {
			log.Fatal("port must be between 1 and 65535")
		}

		if *serverHost == "" {
			log.Fatal("host is required")
		}

		if *serverHttpEndpoint == "" {
			log.Fatal("http-endpoint is required")
		}

		if *serverToken == "" {
			log.Fatal("server token is required")
		}

		startServer(*serverHttpEndpoint, *serverHost, *serverPort, *serverToken)

	case "client":
		err := clientCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}

		if *clientServer == "" {
			log.Fatal("server is required")
		}

		if *clientToken == "" {
			log.Fatal("token is required")
		}

		if *clientForwardUrl == "" {
			log.Fatal("forward-url is required")
		}

		startClient(*clientServer, *clientToken, *clientForwardUrl)

	default:
		println("Starting a webhook relay server: dev-webhook-relay server [options]\nOptions:")
		serverCmd.PrintDefaults()
		println("\nStarting a webhook relay client: dev-webhook-relay client [options]\nOptions:")
		clientCmd.PrintDefaults()
		os.Exit(1)
	}

}
