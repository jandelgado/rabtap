package testcommon

// common integration test functionality. assumes running rabbitmq broker on address
// defined by AMQP_URL and RABBIT_API_URL environment variables.
// (to start a local rabbitmq instance:
//  $ sudo  docker run --rm -ti -p5672:5672 rabbitmq:3-management)

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type FunctionToCapture func(*os.File)

// CaptureOutput captures output of a function that write to the given os.File
func CaptureOutput(f FunctionToCapture) string {
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	out := make(chan string, 1)

	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, reader)
		if err != nil {
			panic(fmt.Sprintf("capturing failed with: %v", err))
		} else {
			out <- buf.String()
		}
	}()

	f(writer)

	_ = writer.Close()

	return <-out
}

// IntegrationAPIURIFromEnv return the REST API URL to use for tests
func IntegrationAPIURIFromEnv() string {
	url := os.Getenv("RABBIT_API_URL")
	if url == "" {
		url = "http://guest:password@localhost:15672/api"
	}
	return url
}

// IntegrationURIFromEnv return the amqp URL to use for tests
func IntegrationURIFromEnv() *url.URL {
	u := os.Getenv("AMQP_URI")
	if u == "" {
		u = "amqp://guest:password@localhost:5672"
	}
	URL, err := url.Parse(u)
	if err != nil {
		panic(err)
	}
	return URL
}

// IntegrationQueueName returns the name of the ith test queue
func IntegrationQueueName(i int) string {
	return fmt.Sprintf("queue-%d", i)
}

// IntegrationTestConnection creates connection to rabbitmq broker and sets up
// optionally an exchange of the given type and bind given number of queues to
// the exchange.  The binding key will aways be the queue name. The queues are
// named "queue-0" "queue-1" etc (see integrationQueueName() func).  If
// parameter addRoutingHeader is true, then the queue will be bound using an
// additional routing header ("x-match":"any", "header1":"test0" for first
// queue etc; this feature is needed by the headers test).
func IntegrationTestConnection(t *testing.T, exchangeName, exchangeType string,
	numQueues int, addRoutingHeader bool,
) (*amqp.Connection, *amqp.Channel) {
	conn, err := amqp.Dial(IntegrationURIFromEnv().String())
	require.Nil(t, err)
	ch, err := conn.Channel()
	require.Nil(t, err)

	if exchangeName == "" {
		return conn, ch
	}
	// create test exchanges and queues
	err = ch.ExchangeDeclare(
		exchangeName,
		exchangeType,
		false, // non durable
		true,  // auto delete
		false, // internal
		false, // wait for response
		nil)
	require.Nil(t, err)

	for i := 0; i < numQueues; i++ {
		queue, err := ch.QueueDeclare(
			IntegrationQueueName(i), // name of the queue
			false,                   // non durable
			false,                   // delete when unused
			true,                    // exclusive
			false,                   // wait for response
			nil)                     // arguments
		require.Nil(t, err)

		// set routing header if requested (used by headers testcase)
		headers := amqp.Table{}
		if addRoutingHeader {
			headerValue := fmt.Sprintf("test%d", i)
			headers = amqp.Table{
				"x-match": "any",
				"header1": headerValue,
			}
		}
		err = ch.QueueBind(
			queue.Name,
			queue.Name,   // bindingKey
			exchangeName, // sourceExchange
			false,        // wait
			headers,
		)
		assert.Nil(t, err)
	}
	return conn, ch
}

// PublishTestMessages publishes the given number of test messages the
// exchange exhangeName with the provided routingKey
func PublishTestMessages(t *testing.T, ch *amqp.Channel, numMessages int,
	exchangeName, routingKey string, optHeaders amqp.Table,
) {
	// inject messages into exchange. Each message should become visible
	// in the tap-exchange defined above.
	for i := 1; i <= numMessages; i++ {
		// publish the test message
		err := ch.PublishWithContext(
			context.TODO(),
			exchangeName,
			routingKey,
			false, // mandatory
			false, // immediate
			amqp.Publishing{
				Body:         []byte("Hello"), // TODO add index
				ContentType:  "text/plain",
				DeliveryMode: amqp.Transient,
				// optional headers only needed in heades test case for routing
				Headers: optHeaders,
			})
		require.Nil(t, err)
	}
}

// VerifyTestMessageOnQueue checks that the expected messages were received on
// the given queue.  on success the number of received messages  is sent
// through the provided success channel signalling success.
func VerifyTestMessageOnQueue(t *testing.T, ch *amqp.Channel, consumer string, numExpected int, queueName string, success chan int) {
	deliveries, err := ch.Consume(
		queueName,
		consumer, // consumer
		false,    // noAck
		true,     // exclusive
		false,    // noLocal
		false,    // noWait
		nil,      // arguments
	)
	require.Nil(t, err)
	numReceived := 0

	go func() {
		for d := range deliveries {
			if string(d.Body) == "Hello" {
				numReceived++
			}

			// Await NumExpectedMessages
			if numReceived == numExpected {
				success <- numReceived
			}
		}
	}()
}
