package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
)

// Global variables
var ctx = context.Background()
var redisClient *redis.Client

// Init function to start redis server
func init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Test the connection
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
}

// Pokemon -> characteristics
type Pokemon struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	XP    int    `json:"xp"`
	Power string `json:"power"`
	Level int    `json:"level"`
}

func getPokemonByType(pokemonType string) ([]Pokemon, error) {
	pokemonRedisPattern := fmt.Sprintf("pokemon:%s:*", pokemonType)

	keys, err := redisClient.Keys(ctx, pokemonRedisPattern).Result()
	if err != nil {
		return nil, err
	}

	var pokemons []Pokemon
	for _, key := range keys {
		// data in redis is stored as json string
		data, err := redisClient.Get(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		var p Pokemon
		// format data to struct
		if err := json.Unmarshal([]byte(data), &p); err != nil {
			return nil, err
		}
		pokemons = append(pokemons, p)
	}

	return pokemons, nil
}

func handlePokemonType(pokemonType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pokemon, err := getPokemonByType(pokemonType)
		if err != nil {
			http.Error(w, "failed to fetch data", http.StatusInternalServerError)
			return
		}

		jsonResponse(w, pokemon)
	}
}

// JSON response encoder
func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode data", http.StatusInternalServerError)
	}
}

func main() {
	routes := map[string]string{
		"/water":     "water",
		"/electric":  "electric",
		"/grass":     "grass",
		"/legendary": "legendary",
		"/fire":      "fire",
	}

	for route, pokemonType := range routes {
		http.HandleFunc(route, handlePokemonType(pokemonType))
	}

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
