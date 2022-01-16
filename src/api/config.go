package api

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type Configuration struct {
	Host                 string `yaml:"host"`
	Port                 uint16 `yaml:"port"`
	OnlineMode           bool   `yaml:"online_mode"`
	MOTD                 string `yaml:"motd"`
	MaxPlayers           int    `yaml:"max_players"`
	Difficulty           string `yaml:"difficulty"`
	CompressionThreshold int32  `yaml:"compression_threshold"`
	Seed                 string `yaml:"seed"`
	Hardcore             bool   `yaml:"hardcore"`
	DefaultGamemode      string `yaml:"default_gamemode"`
	ViewDistance         int    `yaml:"view_distance"`
	SimulationDistance   int    `yaml:"simulation_distance"`
	KeepAliveInterval    int    `yaml:"keep_alive_interval"`
	EnableQuery          bool   `yaml:"enable_query"`
	QueryHost            string `yaml:"query_host"`
	QueryPort            uint16 `yaml:"query_port"`
	Worlds               []struct {
		Name      string `yaml:"name"`
		Generator string `yaml:"generator"`
		Store     struct {
			Name    string                 `yaml:"name"`
			Options map[string]interface{} `yaml:"options"`
		} `yaml:"store"`
	} `yaml:"worlds"`
}

func NewConfiguration() Configuration {
	seed := make([]byte, 16)

	if _, err := rand.Read(seed); err != nil {
		log.Fatal(err)
	}

	return Configuration{
		Host:               "0.0.0.0",
		Port:               25565,
		OnlineMode:         true,
		MOTD:               "A Minecraft Server",
		MaxPlayers:         20,
		Difficulty:         "normal",
		Seed:               hex.EncodeToString(seed),
		Hardcore:           false,
		DefaultGamemode:    "survival",
		ViewDistance:       10,
		SimulationDistance: 10,
		KeepAliveInterval:  15,
		EnableQuery:        true,
		QueryHost:          "0.0.0.0",
		QueryPort:          25565,
	}
}

func (c *Configuration) ReadFromFile(path string) error {
	data, err := ioutil.ReadFile(path)

	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, c)
}

func (c Configuration) WriteToFile(config string) error {
	data, err := yaml.Marshal(c)

	if err != nil {
		return err
	}

	return ioutil.WriteFile(config, data, 0666)
}

func (c Configuration) Validate() error {
	if c.MaxPlayers < 0 {
		return errors.New("config value \"max_players\" may not be a negative value")
	}

	if c.Difficulty != "peaceful" && c.Difficulty != "easy" && c.Difficulty != "normal" && c.Difficulty != "hard" {
		return fmt.Errorf("config value \"difficulty\" has unknown value: %s", c.Difficulty)
	}

	if c.DefaultGamemode != "survival" && c.DefaultGamemode != "creative" && c.DefaultGamemode != "adventure" && c.DefaultGamemode != "spectator" {
		return fmt.Errorf("config value \"default_gamemode\" has unknown value: %s", c.DefaultGamemode)
	}

	// TODO make sure all config properties are validated

	return nil
}
