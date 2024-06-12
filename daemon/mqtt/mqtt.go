package mqtt

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/loopfz/gadgeto/tonic"
	"github.com/sirupsen/logrus"
	"github.com/wI2L/fizz"

	"github.com/rclsilver-org/monitoring/daemon/pkg/component"
	"github.com/rclsilver-org/monitoring/daemon/pkg/server"
)

type Topic struct {
	Timestamp time.Time `json:"timestamp"`
	Payload   []byte    `json:"payload"`
}

type MQTTComponent struct {
	cfg *Config

	health    healthOut
	healthMut sync.Mutex

	topics    map[string]*Topic
	topicsMut sync.Mutex
}

func New(cfg *Config, s *server.Server) (*MQTTComponent, error) {
	mqtt := &MQTTComponent{
		cfg: cfg,

		health: healthOut{
			Status: healthOutStatusUnknown,
		},
	}

	group := s.RegisterGroup("/mqtt", "mqtt", "MQTT monitoring API")

	group.GET("/health", []fizz.OperationOption{
		fizz.Summary("Get the state of the MQTT broker"),
		fizz.Response(fmt.Sprint(http.StatusInternalServerError), "Server Error", server.APIError{}, nil, nil),
	}, tonic.Handler(mqtt.handlerPing, http.StatusOK))

	group.GET("/topics", []fizz.OperationOption{
		fizz.Summary("Get the list of the topics"),
		fizz.Response(fmt.Sprint(http.StatusInternalServerError), "Server Error", server.APIError{}, nil, nil),
	}, tonic.Handler(mqtt.handlerListTopics, http.StatusOK))

	group.GET("/topics/:topic", []fizz.OperationOption{
		fizz.Summary("Get the details of a topic"),
		fizz.Response(fmt.Sprint(http.StatusInternalServerError), "Server Error", server.APIError{}, nil, nil),
	}, tonic.Handler(mqtt.handlerGetTopic, http.StatusOK))

	return mqtt, nil
}

func (c *MQTTComponent) setAvailable() {
	c.healthMut.Lock()
	defer c.healthMut.Unlock()

	c.health.Status = healthOutStatusOK
	c.health.Error = nil
}

func (c *MQTTComponent) setUnavailable(err error) {
	c.healthMut.Lock()
	defer c.healthMut.Unlock()

	c.health.Status = healthOutStatusOK
	c.health.Error = err
}

func (c *MQTTComponent) disconnect() {
	c.healthMut.Lock()
	defer c.healthMut.Unlock()

	c.health.Status = healthOutStatusUnknown
	c.health.Error = nil

	logrus.Debug("connection closed to the MQTT broker")
}

func (c *MQTTComponent) Run(ctx context.Context) error {
	logrus.WithContext(ctx).Debug("starting the MQTT component")

	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("unable to get current hostname: %w", err)
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", c.cfg.Host, c.cfg.Port))
	opts.SetClientID(fmt.Sprintf("mqtt-monitoring-%s", hostname))
	//opts.SetUsername("emqx")
	//opts.SetPassword("public")

	// auto-connect
	opts.ConnectTimeout = time.Second * 10
	opts.AutoReconnect = true
	opts.ConnectRetry = true
	opts.ConnectRetryInterval = time.Second * 5

	// handlers
	opts.OnConnect = c.connectHandler
	opts.OnConnectionLost = c.connectLostHandler
	opts.OnReconnecting = c.reconnectingHandler

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		err := fmt.Errorf("error while connecting: %w", token.Error())
		c.setUnavailable(err)
		return err
	}

	if token := client.Subscribe("#", 0, c.messagePubHandler); token.Wait() && token.Error() != nil {
		err := fmt.Errorf("error while subscribing: %w", token.Error())
		c.disconnect()
		return err
	}

	select {
	case <-ctx.Done():
		c.disconnect()
		return component.ErrInterrupted
	}
}

func (c *MQTTComponent) messagePubHandler(client mqtt.Client, msg mqtt.Message) {
	c.topicsMut.Lock()
	defer c.topicsMut.Unlock()

	if c.topics == nil {
		c.topics = make(map[string]*Topic)
	}

	_, ok := c.topics[msg.Topic()]

	if !ok {
		c.topics[msg.Topic()] = &Topic{
			Timestamp: time.Unix(0, 0),
		}
	}

	if ok || !msg.Retained() {
		c.topics[msg.Topic()].Timestamp = time.Now()
	}

	c.topics[msg.Topic()].Payload = make([]byte, 0, len(msg.Payload()))
	c.topics[msg.Topic()].Payload = append(c.topics[msg.Topic()].Payload, msg.Payload()...)
}

func (c *MQTTComponent) connectHandler(client mqtt.Client) {
	c.setAvailable()
	logrus.Infof("connected to the MQTT broker: %s:%d", c.cfg.Host, c.cfg.Port)
}

func (c *MQTTComponent) reconnectingHandler(client mqtt.Client, opts *mqtt.ClientOptions) {
	c.setUnavailable(fmt.Errorf("reconnecting"))
	logrus.Warning("reconnecting to the MQTT broker")
}

func (c *MQTTComponent) connectLostHandler(client mqtt.Client, err error) {
	c.setUnavailable(err)
	logrus.WithError(err).Error("connection lost from the MQTT broker")
}
