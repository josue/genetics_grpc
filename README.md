### gRPC Challenge

1.  Company routinely transfers data from laboratory experiments to our cloud-based data warehouse.

    a. Commonly, our experiments are conducted in 96-well plates that are handled by robots. These plates are depicted below, with A-H on the rows and 1-12 on the columns. Wells are named A1-H12 based on the intersection of rows and columns (see image below).

    b. Each robot handles multiple plates simultaneously in a run (i.e., batch), and data is transferred to the data warehouse on a per-run basis.

    c. For each well, optical density (a proxy for cell growth) is collected as a time-series during the run (typically 1000-2000 measurements per well over a few days). Optical density typically ranges from 0.1 to 2.0. At the end of the run, the robot generates a comma separated file with the following columns:

        i. Time (in whole seconds) since the start of the run.
        ii. Well name
        iii. Optical density
        iv. Corrected optical density

    Using protobuf and gRPC (or REST API, not preferred):

    a) Create a command-line based client that the robot can execute to send the parsed data.

    b) Create a server that prints out the data in a structured way.

    c) Bonus: Instead of printing the data as in b), connect to a SQL database (preferably postgresql) and deposit the data in an appropriately structured schema (note: database knowledge is not a requirement for this position, but is a plus)

2.  Create a Docker container for the service you’ve developed above.

    a. Include the Dockerfile

    b. Describe the steps you would take to deploy an updated image when triggered by a git push to a master branch.

---

### Solution:

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

#### Deploy Image Update Procedure

> Describe the steps you would take to deploy an updated image when triggered by a git push to a master branch.

1. Setup a Cloudformation stack with a CodeBuild build specs file + CodePipeline.
2. Upon a git push to master branch on Github, CodeBuild will capture the Github webhook updates.
3. CodeBuild creates the container image from the buildspec instructions and the final image is pushed to Dockerhub.
4. Then CodePipeline is configured to deploy the image to an existing ECS/Fargate deployment using a TaskDefinition file.
