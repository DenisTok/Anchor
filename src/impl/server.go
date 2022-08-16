package impl

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path"

	"github.com/anchormc/anchor/src/api"
	"github.com/anchormc/anchor/src/api/conf"
	gameAPI "github.com/anchormc/anchor/src/api/game"
	"github.com/anchormc/anchor/src/api/log"
	netAPI "github.com/anchormc/anchor/src/api/net"
	"github.com/anchormc/anchor/src/impl/game"
	"github.com/anchormc/anchor/src/impl/net"
)

type Server struct {
	isRunning bool
	cwd       string
	config    *conf.Configuration
	socket    netAPI.Socket
	worlds    map[string]gameAPI.World
	clients   []api.Client
	favicon   image.Image
}

func NewServer() api.Server {
	return &Server{
		config:  conf.NewConfiguration(),
		socket:  net.NewSocket(),
		clients: make([]api.Client, 0),
		worlds:  make(map[string]gameAPI.World),
	}
}

func (s *Server) Initialize() error {
	cwd, err := os.Getwd()

	if err != nil {
		return err
	}

	s.cwd = cwd

	if _, err = os.Stat(path.Join(cwd, "config.yml")); err == nil {
		if err = s.config.ReadFile(path.Join(cwd, "config.yml")); err != nil {
			return err
		}

		if err = s.config.Validate(); err != nil {
			return err
		}

		log.SetLogLevel(s.config.LogLevel)
		log.Info("Successfully loaded configuration file")
	}

	if err = s.config.WriteFile(path.Join(cwd, "config.yml")); err != nil {
		return err
	}

	if _, err = os.Stat(path.Join(cwd, "server-icon.png")); err == nil {
		f, err := os.Open(path.Join(cwd, "server-icon.png"))

		if err != nil {
			return err
		}

		defer f.Close()

		img, err := png.Decode(f)

		if err != nil {
			return err
		}

		bounds := img.Bounds()

		if bounds.Max.X != 64 || bounds.Max.Y != 64 {
			return fmt.Errorf("invalid server-icon.png dimensions, expected 64x64, got %dx%d", bounds.Max.X, bounds.Max.Y)
		}

		s.favicon = img

		log.Info("Loaded server icon image")
	}

	if _, err := s.CreateWorld("world", nil); err != nil {
		return err
	}

	return nil
}

func (s *Server) Start() error {
	if s.isRunning {
		return errors.New("attempted to Start() server when it is already running")
	}

	if err := s.socket.Start(s.config.Host, s.config.Port); err != nil {
		return err
	}

	s.isRunning = true

	log.Infof("Listening for connections on %s:%d\n", s.config.Host, s.config.Port)

	go s.AcceptConnections()

	return nil
}

func (s *Server) AcceptConnections() {
	for s.isRunning && s.socket.IsRunning() {
		conn, err := s.socket.OnConnection()

		if err != nil {
			if !s.isRunning {
				break
			}

			log.Error(err)

			continue
		}

		client, err := net.NewClient(conn)

		if err != nil {
			log.Error(err)

			if err = conn.Close(); err != nil {
				log.Error(err)
			}

			continue
		}

		s.clients = append(s.clients, client)

		log.Infof("Received a connection from %s (connected: %d)\n", conn.RemoteAddr(), len(s.clients))

		go client.HandlePackets(context.TODO(), s)
	}
}

func (s Server) Favicon() image.Image {
	return s.favicon
}

func (s Server) GetAllClients() []api.Client {
	return s.clients
}

func (s *Server) RemoveClient(uuid string) {
	for k, c := range s.clients {
		if c.UUID() != uuid {
			continue
		}

		s.clients = append(s.clients[0:k], s.clients[k+1:]...)
	}
}

func (s Server) GetConfig() *conf.Configuration {
	return s.config
}

func (s Server) PlayerCount() (count int) {
	for _, client := range s.clients {
		if client.GetPlayer() == nil {
			continue
		}

		count++
	}

	return
}

func (s Server) GetWorld(name string) gameAPI.World {
	return s.worlds[name]
}

func (s *Server) CreateWorld(name string, generator gameAPI.Generator) (gameAPI.World, error) {
	cache, ok := s.worlds[name]

	if ok {
		return cache, nil
	}

	folder := path.Join(s.cwd, "worlds", name)

	if err := os.MkdirAll(folder, 0777); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(path.Join(folder, "region"), 0777); err != nil {
		return nil, err
	}

	world := game.NewWorld(name, folder, generator)

	if err := world.LoadAllRegions(); err != nil {
		return nil, err
	}

	s.worlds[name] = world

	return world, nil
}

func (s *Server) Close() error {
	if !s.isRunning {
		return errors.New("attempted to Close() server when it is not running")
	}

	// TODO server closing duties

	s.isRunning = false

	if err := s.socket.Close(); err != nil {
		return err
	}

	log.Info("Server closed")

	return nil
}

var _ api.Server = &Server{}
