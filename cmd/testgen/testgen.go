package main

// testgen is a simple test data generator, which sets up a topology and
// publishes test-messages. it is used for manual tests. see code for
// details.
//
// Usage:
// ./testgen [-numq 5] [-delay 1]
//
// Specify broker with RABTAP_TESTGEN_AMQP_URI environment variable. Default is
// amqp://guest:guest@127.0.0.1:5672

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

const (
	testExchangeDirect  = "test-direct"
	testExchangeFanout  = "test-fanout"
	testExchangeTopic   = "test-topic"
	testExchangeHeaders = "test-headers"
)

func rabbitURLFromEnv() string {
	url := os.Getenv("RABTAP_TESTGEN_AMQP_URI")
	if url == "" {
		url = "amqp://guest:guest@127.0.0.1:5672"
	}
	return url
}

func failOnError(err error, msg string) {
	if err != nil {
		//	log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func createExchange(ch *amqp.Channel,
	exchangeName, exchangeType string) {

	// create test exchanges and queues
	err := ch.ExchangeDeclare(
		exchangeName,
		exchangeType,
		false, // non durable
		true,  // auto delete
		false, // internal
		true,  // wait for response
		nil)
	failOnError(err, "could not declare exchange "+exchangeName)
}

func createQueue(ch *amqp.Channel, queueName string) {
	_, err := ch.QueueDeclare(
		queueName,
		false, // non durable
		false, // delete when unused
		false, // non exclusive
		true,  // wait for response
		nil)   // arguments
	failOnError(err, "could not create queue "+queueName)
}

func bindQueue(ch *amqp.Channel, queueName, exchangeName, bindingKey string,
	headers amqp.Table) {
	err := ch.QueueBind(
		queueName,
		bindingKey,
		exchangeName,
		true, // wait
		headers,
	)
	failOnError(err, "could not bind queue "+queueName+
		" to exchange "+exchangeName)
}

func getRoutingKeyForExchange(exchangeName string, i int) (string, amqp.Table) {

	switch exchangeName {
	case "amq.headers":
		fallthrough
	case testExchangeHeaders:
		// a header exchange needs a routing header hash map.
		return "", amqp.Table{
			"header1": fmt.Sprintf("test%d", i),
		}
	default:
		return getQueueNameForExchange(exchangeName, i), amqp.Table{}
	}

}

func getBindingKeyForExchange(exchangeName string, i int) (string, amqp.Table) {

	switch exchangeName {
	case "amq.headers":
		fallthrough
	case testExchangeHeaders:
		// a header exchange needs set a routing header hash map.
		return "", amqp.Table{
			"x-match": "all",
			"header1": fmt.Sprintf("test%d", i),
		}

	default:
		return getQueueNameForExchange(exchangeName, i), amqp.Table{}
	}

}

func getQueueNameForExchange(exchangeName string, i int) string {
	return fmt.Sprintf("test-q-%s-%d", exchangeName, i)
}

func generateTestMessages(ch *amqp.Channel, exchanges []string, numTestQueues int, delay time.Duration) {
	count := 1
	for {
		for _, exchange := range exchanges {
			for i := 0; i < numTestQueues; i++ {
				routingKey, headers := getRoutingKeyForExchange(exchange, i)
				log.Printf("publishing msg #%d to exchange '%s' with routing key '%s' and headers %#+v",
					count, exchange, routingKey, headers)
				err := ch.Publish(
					exchange,
					routingKey,
					false, // mandatory
					false, // immediate
					amqp.Publishing{
						Body: []byte(fmt.Sprintf(
							"test message #%d was pushed to exchange '%s' with routing key '%s' and headers %#+v",
							count, exchange, routingKey, headers)),
						ContentType:  "text/plain",
						AppId:        "rabtap.testgen",
						Timestamp:    time.Now(),
						DeliveryMode: amqp.Transient,
						Headers:      headers,
					})
				failOnError(err, "publish failed")
				count++
			}
		}
		time.Sleep(delay)
	}
}

func createTopology(ch *amqp.Channel, exchanges []string, numTestQueues int) {
	// create queues and bindings. Binding key = queue name
	for _, exchange := range exchanges {
		for i := 0; i < numTestQueues; i++ {
			queueName := getQueueNameForExchange(exchange, i)
			log.Printf("creating queue %s", queueName)
			createQueue(ch, queueName)
			bindingKey, headers := getBindingKeyForExchange(exchange, i)
			log.Printf("binding queue %s to exchange %s with bindingkey `%s` and headers %#+v",
				queueName, exchange, bindingKey, headers)
			bindQueue(ch, queueName, exchange, bindingKey, headers)
		}
	}
}

func main() {

	numTestQueues := flag.Int("numq", 5, "number of queues to create")
	delayTime := flag.Int("delay", 1, "delay in s between sending of message chunks")
	flag.Parse()

	log.Printf("will create %d queues", *numTestQueues)
	log.Printf("connecting to %s", rabbitURLFromEnv())
	conn, err := amqp.Dial(rabbitURLFromEnv())
	failOnError(err, "could no connect to broker")

	ch, err := conn.Channel()
	failOnError(err, "could no create channel")

	log.Printf("creating exchanges and queues")
	createExchange(ch, testExchangeDirect, amqp.ExchangeDirect)
	createExchange(ch, testExchangeFanout, amqp.ExchangeFanout)
	createExchange(ch, testExchangeTopic, amqp.ExchangeTopic)
	createExchange(ch, testExchangeHeaders, amqp.ExchangeHeaders)

	exchanges := []string{
		"amq.direct",
		"amq.topic",
		"amq.fanout",
		"amq.headers",
		testExchangeDirect,
		testExchangeFanout,
		testExchangeTopic,
		testExchangeHeaders}

	createTopology(ch, exchanges, *numTestQueues)
	generateTestMessages(ch, exchanges, *numTestQueues, time.Duration(*delayTime)*time.Second)
}
