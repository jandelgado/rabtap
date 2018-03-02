# testgen - rabtap test data generator

## Usage
```
Usage of ./testgen:
  -delay int
    	delay in s between sending of message chunks (default 1)
  -numq int
    	number of queues to create (default 5)
```

Use the `RABTAP_TESTGEN_AMQP_URI` environment variable to specify the RabbitMQ
broker to be used. If not set, `amqp://guest:guest@localhost:5672` will be used.

Be careful when running this tool in a non-test environment since it creates
exchanges, queues and publishes messages.

## Build

Just run `make` to build testgen.
