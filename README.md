### gRPC stream with client & server

Demonstrates a gRPC client parsing a large CSV formatted dataset file and quickly streams all records to a gRPC server with confirmation received, and inserts all records into a Postgres database.

Requires: **Docker** + **Golang**

#### Makefile targets for running programs:

Run Golang programs, Docker and Docker-Compose setup via the follow targets:

- `make generate_protobuf` - Generates the protobuf Go package into client + server internal directories

- `make stack_up` - Executes **docker-compose up -d** command

- `make stack_down` - Executes **docker-compose down** command

- `make stack_logs` - Executes **docker-compose logs -f** command

- `make docker_db` - Starts a Postgres docker container.

- `make run_server_with_stdout` - Runs the server Golang main file with STDOUT output.

- `make run_server_with_db` - Runs the server Golang main file with database insert only, no output.

- `make run_client` - Runs the client Golang main file

#### Important Testing/Development Note:

It's highly recommended to run the `stack_up` target as this will build, install & run the Golang apps along with the database inside a docker-compose environment.

_See the [docker-compose yaml](docker-compose.yml) file for details._

Expected stack logs output (in no particular order/format):

```
db_1      | PostgreSQL Database directory appears to contain a database; Skipping initialization
db_1      | 2020-05-26 06:54:29.010 UTC [1] LOG:  starting PostgreSQL 12.3 (Debian 12.3-1.pgdg100+1) ...
db_1      | 2020-05-26 06:54:29.043 UTC [1] LOG:  database system is ready to accept connections

server_1  | 2020/05/26 06:54:29 Output type: db
server_1  | 2020/05/26 06:54:29 DB Init - Table exists
server_1  | 2020/05/26 06:54:29 gRPC server starting on 50051 ...
server_1  | 2020/05/26 06:54:32 Received 93504 Plates, Now Processing ...
server_1  | 2020/05/26 06:54:32 Saving plates received to database table: plates ...

client_1  | 2020/05/26 06:54:29 Reading file: test/od-data.csv
client_1  | 2020/05/26 06:54:29 Client sending 93504 Plates
client_1  | 2020/05/26 06:54:32 Server response: Received 93504 Plates, Now Processing ...
client_1  | 2020/05/26 06:54:32 Duration: 2.4904113s
```

If you want to run individually, then you must run the targets in this order:

1. `make docker_db`
2. `make run_server_with_db`
3. `make run_client`

