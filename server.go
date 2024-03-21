package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// The currently connected websocket client.
var client *websocket.Conn = nil

// Used to upgrade websocket connections.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WebsocketRequestHandler func(w http.ResponseWriter, r *http.Request)

func createWebsocketHandler(token string) WebsocketRequestHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		// If a client is already connected, reject the new connection.
		if client != nil {
			_ = client.Close()
			client = nil
		}

		// Check for authorization header.
		authHeader := r.Header.Get("Authorization")

		if authHeader != fmt.Sprintf("Bearer %s", token) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Upgrade initial GET request to a WebSocket connection.
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Error upgrading to WebSocket: %v\n", err)
			return
		}

		// Set the client to the new WebSocket connection.
		log.Printf("Client connected from %s\n", r.RemoteAddr)
		client = ws
	}
}

// handleHTTPForward forwards HTTP requests to WebSocket clients.
func handleHTTPForward(w http.ResponseWriter, r *http.Request) {
	if client == nil {
		http.Error(w, "No client connected", http.StatusInternalServerError)
		return
	}

	// Read the request data into a struct.
	requestData := RequestData{
		Method:  r.Method,
		Headers: r.Header,
	}

	log.Printf("Got %s request. Forwarding to connected client.\n", requestData.Method)

	// Read the body, if any.
	// TODO: Instead of reading all at once, stream over to the client.
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	requestData.Body = string(bodyBytes)

	// Serialize the requestData struct to JSON
	jsonData, err := json.Marshal(requestData)

	if err != nil {
		http.Error(w, "Error encoding request data", http.StatusInternalServerError)
		return
	}

	// Send the JSON data to the client and read the response.
	err = client.WriteMessage(websocket.TextMessage, jsonData)

	if err != nil {
		http.Error(w, "Error sending request data", http.StatusInternalServerError)
		_ = client.Close()
		client = nil
		return
	}

	// Read the response from the client.
	messageType, response, err := client.ReadMessage()

	if err != nil || messageType != websocket.TextMessage {
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		_ = client.Close()
		client = nil
		return
	}

	// Parse the response into a ResponseData struct.
	var responseData ResponseData
	err = json.Unmarshal(response, &responseData)

	if err != nil {
		http.Error(w, "Error decoding response data", http.StatusInternalServerError)
		return
	}

	// Set the response headers and status code.
	for key, value := range responseData.Headers {
		w.Header()[key] = value
	}

	w.WriteHeader(responseData.StatusCode)

	// Write the response body.
	if len(responseData.Body) > 0 {
		_, err = w.Write([]byte(responseData.Body))
		if err != nil {
			http.Error(w, "Error writing response body", http.StatusInternalServerError)
			return
		}
	}
}

func startServer(httpEndpoint string, bindHost string, bindPort uint, token string) {
	// Set up our HTTP and WebSocket endpoints.
	http.HandleFunc(httpEndpoint, handleHTTPForward)
	http.HandleFunc("/ws", createWebsocketHandler(token))

	log.Printf("Starting server on %s:%d", bindHost, bindPort)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", bindHost, bindPort), nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
