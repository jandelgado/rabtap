Configuration for the RabbitMQ broker as used in the rabtap integration tests:

The broker offers the following services:
 * 15671/15672 - Management interface (https/http) 
 * 5671/5672 - AMQP (TLS/non TLS)

The server is configured for PLAIN an EXTERNAL (mTLS) auth on the AMQP endpoint.

Certificates are expected to be in the `/certs` directory, and are created
by the `mkcerts.sh` script.

User configured:
 * guest/password  -> used with PLAIN auth
 * testuser -> used with client certificates

