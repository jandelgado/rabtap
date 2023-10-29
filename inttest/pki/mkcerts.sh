#!/bin/bash
# create a CA, server and client certificates which are used by RabbitMQ
# and some of the integration tests. Certificates are created inside the 
# the ./certs directory
set -eou pipefail

CA_CN=${CA_CN:-rabtap testing CA}
SERVER_CN=${SERVER_CN:-localhost}
CLIENTS=${CLIENTS:-testuser unknown guest}

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

mkdir -p "$DIR/certs" && cd certs

function make_ca {

    # Generate self signed root CA cert => ca.key, ca.crt
    openssl req -nodes -x509 -newkey rsa:2048 \
                -keyout ca.key -out ca.crt \
                -subj "/C=DE/ST=HESSEN/L=Frankfurt/O=Rabtap/OU=root/CN=$CA_CN/emailAddress=me@example.com"

    # Generate server cert to be signed => server.key, server.csr
    openssl req -new -nodes -newkey rsa:2048 \
                -keyout server.key -out server.csr \
                -subj "/C=DE/ST=HESSEN/L=Frankfurt/O=Rabtap/OU=server/CN=$SERVER_CN/emailAddress=me@example.com/"
    #            -addext "subjectAltName = DNS:rabbitmq.local"
        chmod 644 ca.key
}

function make_server {
    # Sign the server cert => server.crt. Note that the SAN records
    # (Subject Alternate Names) are set here, not in the CSR.
    openssl x509 -req  -in server.csr -CA ca.crt\
                 -CAkey ca.key -CAcreateserial -out server.crt\
                 -extfile <(printf "[SAN]\nsubjectAltName=DNS:$SERVER_CN")\
                 -extensions SAN
        chmod 644 server.key
}

function make_clients {
    for client in $CLIENTS; do 

        # Generate client cert to be signed => client.key, client.csr
        openssl req -nodes -newkey rsa:2048 \
                    -keyout ${client}.key -out ${client}.csr \
                    -subj "/C=DE/ST=HESSEN/L=Frankfurt/O=Rabtap/OU=${client}/CN=${client}/emailAddress=me@example.com"

        # Sign the client cert => client.crt
        openssl x509 -req -in ${client}.csr \
                     -days 36000 \
                     -CA ca.crt -CAkey ca.key -CAserial ca.srl \
                     -out ${client}.crt
        chmod 644 ${client}.key
    done
}

make_ca
make_server
make_clients

