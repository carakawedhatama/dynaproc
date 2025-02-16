package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAMQPChannel struct {
	mock.Mock
}

func (m *MockAMQPChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	called := m.Called(name, durable, autoDelete, exclusive, noWait, args)
	return called.Get(0).(amqp.Queue), called.Error(1)
}

func (m *MockAMQPChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	args := m.Called(exchange, key, mandatory, immediate, msg)
	return args.Error(0)
}

func (m *MockAMQPChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	called := m.Called(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	return called.Get(0).(<-chan amqp.Delivery), called.Error(1)
}

type MockAMQPConnection struct {
	mock.Mock
}

func (m *MockAMQPConnection) Channel() (*amqp.Channel, error) {
	args := m.Called()
	return args.Get(0).(*amqp.Channel), args.Error(1)
}

func (m *MockAMQPConnection) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestPublishToQueue(t *testing.T) {
	// Test case 1: Successful publish
	t.Run("Successfully publish order", func(t *testing.T) {
		mockChannel := new(MockAMQPChannel)

		originalChannel := rabbitChannel
		defer func() { rabbitChannel = originalChannel }()
		rabbitChannel = mockChannel

		mockChannel.On("QueueDeclare",
			"purchase_orders", // name
			true,              // durable
			false,             // autoDelete
			false,             // exclusive
			false,             // noWait
			amqp.Table(nil),   // args
		).Return(amqp.Queue{Name: "purchase_orders"}, nil)

		order := PurchaseOrder{
			ID:       "PO123",
			VendorID: "V001",
			Amount:   100.50,
			Currency: "USD",
		}

		expectedBody, _ := json.Marshal(order)
		mockChannel.On("Publish",
			"",                // exchange
			"purchase_orders", // routing key
			false,             // mandatory
			false,             // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        expectedBody,
			},
		).Return(nil)

		err := PublishToQueue(order)
		assert.NoError(t, err)

		mockChannel.AssertExpectations(t)
	})

	// Test case 2: Queue declare error
	t.Run("Queue declare error", func(t *testing.T) {
		mockChannel := new(MockAMQPChannel)
		originalChannel := rabbitChannel
		defer func() { rabbitChannel = originalChannel }()
		rabbitChannel = mockChannel

		mockChannel.On("QueueDeclare",
			"purchase_orders",
			true,
			false,
			false,
			false,
			amqp.Table(nil),
		).Return(amqp.Queue{}, errors.New("queue declare error"))

		order := PurchaseOrder{
			ID:       "PO123",
			VendorID: "V001",
			Amount:   100.50,
			Currency: "USD",
		}

		err := PublishToQueue(order)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "queue declare error")
	})

	// Test case 3: Publish error
	t.Run("Publish error", func(t *testing.T) {
		mockChannel := new(MockAMQPChannel)
		originalChannel := rabbitChannel
		defer func() { rabbitChannel = originalChannel }()
		rabbitChannel = mockChannel

		mockChannel.On("QueueDeclare",
			"purchase_orders",
			true,
			false,
			false,
			false,
			amqp.Table(nil),
		).Return(amqp.Queue{Name: "purchase_orders"}, nil)

		mockChannel.On("Publish",
			"",
			"purchase_orders",
			false,
			false,
			mock.Anything,
		).Return(errors.New("publish error"))

		order := PurchaseOrder{
			ID:       "PO123",
			VendorID: "V001",
			Amount:   100.50,
			Currency: "USD",
		}

		err := PublishToQueue(order)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "publish error")
	})
}

func TestConsumeQueue(t *testing.T) {
	t.Run("Successfully consume messages", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			w.WriteHeader(http.StatusOK)
		}))
		defer func() {
			time.Sleep(100 * time.Millisecond)
			server.Close()
		}()

		mockChannel := new(MockAMQPChannel)
		originalChannel := rabbitChannel
		defer func() { rabbitChannel = originalChannel }()
		rabbitChannel = mockChannel

		deliveries := make(chan amqp.Delivery)

		mockChannel.On("QueueDeclare",
			"purchase_orders",
			true,
			false,
			false,
			false,
			amqp.Table(nil),
		).Return(amqp.Queue{Name: "purchase_orders"}, nil)

		mockChannel.On("Consume",
			"purchase_orders",
			"",
			true,
			false,
			false,
			false,
			amqp.Table(nil),
		).Return((<-chan amqp.Delivery)(deliveries), nil)

		cfg := Config{
			RabbitMQ: RabbitMQConfig{
				URL: "amqp://guest:guest@localhost:5672/",
			},
			Dynamics365: Dynamics365Config{
				APIURL: server.URL,
			},
		}

		done := make(chan bool)
		go func() {
			ConsumeQueue(cfg)
			done <- true
		}()

		order := PurchaseOrder{
			ID:       "PO123",
			VendorID: "V001",
			Amount:   100.50,
			Currency: "USD",
		}
		body, _ := json.Marshal(order)
		deliveries <- amqp.Delivery{Body: body}

		close(deliveries)

		<-done

		mockChannel.AssertExpectations(t)
	})

	t.Run("Queue declare error", func(t *testing.T) {
		var buf bytes.Buffer
		originalOutput := log.Writer()
		log.SetOutput(&buf)
		defer func() {
			log.SetOutput(originalOutput)
		}()

		exitCalled := false
		osExit = func(code int) {
			exitCalled = true
			assert.Equal(t, 1, code)
		}

		mockChannel := new(MockAMQPChannel)
		originalChannel := rabbitChannel
		defer func() { rabbitChannel = originalChannel }()
		rabbitChannel = mockChannel

		mockChannel.On("QueueDeclare",
			"purchase_orders",
			true,
			false,
			false,
			false,
			amqp.Table(nil),
		).Return(amqp.Queue{}, errors.New("queue declare error"))

		cfg := Config{
			RabbitMQ: RabbitMQConfig{
				URL: "amqp://guest:guest@localhost:5672/",
			},
		}

		ConsumeQueue(cfg)

		assert.Contains(t, buf.String(), "queue declare error")
		assert.True(t, exitCalled, "os.Exit was not called")
	})

	t.Run("Sync error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer func() {
			time.Sleep(100 * time.Millisecond)
			server.Close()
		}()

		mockChannel := new(MockAMQPChannel)
		originalChannel := rabbitChannel
		defer func() { rabbitChannel = originalChannel }()
		rabbitChannel = mockChannel

		deliveries := make(chan amqp.Delivery)

		mockChannel.On("QueueDeclare",
			"purchase_orders",
			true,
			false,
			false,
			false,
			amqp.Table(nil),
		).Return(amqp.Queue{Name: "purchase_orders"}, nil)

		mockChannel.On("Consume",
			"purchase_orders",
			"",
			true,
			false,
			false,
			false,
			amqp.Table(nil),
		).Return((<-chan amqp.Delivery)(deliveries), nil)

		cfg := Config{
			RabbitMQ: RabbitMQConfig{
				URL: "amqp://guest:guest@localhost:5672/",
			},
			Dynamics365: Dynamics365Config{
				APIURL: server.URL,
			},
		}

		done := make(chan bool)
		go func() {
			ConsumeQueue(cfg)
			done <- true
		}()

		order := PurchaseOrder{
			ID:       "PO123",
			VendorID: "V001",
			Amount:   100.50,
			Currency: "USD",
		}
		body, _ := json.Marshal(order)
		deliveries <- amqp.Delivery{Body: body}
		close(deliveries)

		<-done

		mockChannel.AssertExpectations(t)
	})
}
