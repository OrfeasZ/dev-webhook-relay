package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"strings"
)

func startClient(server string, token string, forwardUrl string) {
	// Create a websocket connection to the server with the provided token.
	authHeader := fmt.Sprintf("Bearer %s", token)
	client, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("%s/ws", server), http.Header{"Authorization": []string{authHeader}})
	if err != nil {
		log.Fatalf("Error connecting to WS server: %v", err)
	}

	defer client.Close()

	log.Printf("Connected to %s\n", server)

	// Wait for incoming messages from the server and forward them as HTTP requests
	// to the provided URL.
	for {
		// Read the message from the server.
		messageType, message, err := client.ReadMessage()
		if err != nil || messageType != websocket.TextMessage {
			log.Fatalf("Error reading WS message: %v", err)
		}

		// Parse the message into a RequestData struct.
		var requestData RequestData
		err = json.Unmarshal(message, &requestData)
		if err != nil {
			log.Fatalf("Error decoding WS message: %v", err)
		}

		// Create a new HTTP request from the RequestData struct.
		log.Printf("Forwarding %s request to %s\n", requestData.Method, forwardUrl)
		req, err := http.NewRequest(requestData.Method, forwardUrl, nil)
		if err != nil {
			log.Fatalf("Error creating HTTP request: %v", err)
		}

		// Set the headers on the request.
		for key, values := range requestData.Headers {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// Set the body on the request.
		req.Body = io.NopCloser(strings.NewReader(requestData.Body))

		// Default response.
		responseData := ResponseData{
			StatusCode: 500,
			Headers:    make(map[string][]string),
			Body:       "Internal server error",
		}

		// Send the request and read the response.
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Error forwarding HTTP request: %v\n", err)
		} else {
			// Read the response body.
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Error reading response body: %v\n", err)
			} else {
				// Create a new ResponseData struct from the response.
				responseData = ResponseData{
					StatusCode: resp.StatusCode,
					Headers:    resp.Header,
					Body:       string(bodyBytes),
				}
			}
		}

		// Serialize the ResponseData struct to JSON.
		jsonData, err := json.Marshal(responseData)
		if err != nil {
			log.Fatalf("Error encoding response data: %v", err)
		}

		// Send the JSON data to the server.
		err = client.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			log.Fatalf("Error sending response data to WS: %v", err)
		}
	}
}
