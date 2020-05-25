package app

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	pb "./proto"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Postgres DB defaults
const (
	defaultDBHost     = "localhost"
	defaultDBName     = "public"
	defaultDBUsername = "postgres"
	defaultDBPassword = "postgres"
	defaultDBPort     = 5432
	defaultDBTable    = "plates"
)

// table schema for Postgres DB
var dbSchema = `
CREATE TABLE plates (
    plate integer NULL,
	well text NULL,
	runtime integer NULL,
	optical_density real NULL,
	run text NULL,
    corrected_optical_density real NULL
)`

// Server struct params and methods
type Server struct {
	pb.PlatesServer
	dbInit      bool
	dbConn      *sqlx.DB
	queuePlates chan pb.PlateRequest

	OutputType string
	DBHost     string
	DBName     string
	DBUsername string
	DBPassword string
	DBPort     int
	DBTable    string
}

// Init method evals the database connection params and starts a queue to work on received plates
func (s *Server) Init() {
	// checks DB params not empty
	if s.DBHost == "" {
		s.DBHost = defaultDBHost
	}
	if s.DBName == "" {
		s.DBName = defaultDBName
	}
	if s.DBUsername == "" {
		s.DBUsername = defaultDBUsername
	}
	if s.DBPassword == "" {
		s.DBPassword = defaultDBPassword
	}
	if s.DBPort == 0 {
		s.DBPort = defaultDBPort
	}
	if s.DBTable == "" {
		s.DBTable = defaultDBTable
	}

	// initialize worker channel/goroutine to accept plate data received from clients for processing
	s.queuePlates = make(chan pb.PlateRequest)
	go s.plateWorker()
}

// GetEnvConfig method evals database conneciton params found via Environment variables
func (s *Server) GetEnvConfig() {
	if os.Getenv("DB_HOST") != "" {
		s.DBHost = os.Getenv("DB_HOST")
	}
	if os.Getenv("DB_NAME") != "" {
		s.DBName = os.Getenv("DB_NAME")
	}
	if os.Getenv("DB_USERNAME") != "" {
		s.DBUsername = os.Getenv("DB_USERNAME")
	}
	if os.Getenv("DB_PASSWORD") != "" {
		s.DBPassword = os.Getenv("DB_PASSWORD")
	}
	if os.Getenv("DB_PORT") != "" {
		num, _ := strconv.Atoi(os.Getenv("DB_PORT"))
		s.DBPort = num
	}
	if os.Getenv("DB_TABLE") != "" {
		s.DBPassword = os.Getenv("DB_TABLE")
	}
}

// SendPlates method accepts a stream from gRPC clients with the plate data to process then it will handoff each plate
// to the worker channel for processing (stdout or save to DB) as non-blocking and will respond back to clients with
// total received plates for confirmation.
func (s *Server) SendPlates(stream pb.Plates_SendPlatesServer) error {
	var records int
	pretext := "Received %v Plates, Now Processing ..."
	message := fmt.Sprintf(pretext, records)

	for {
		pr, err := stream.Recv()

		if err == io.EOF {
			message = fmt.Sprintf(pretext, records)
			stream.SendAndClose(&pb.PlateResponse{Message: message})
			break
		}
		if err != nil {
			return err
		}

		// handoff plate data to plates queue without blocking response to client (especially for large streams)
		// thus sends confirmation quickly while channel is processing in the background each request.
		go func() { s.queuePlates <- *pr }()
		records++
	}

	log.Println(message)

	if s.OutputType == "db" {
		log.Printf("Saving plates received to database table: %v ...\n", s.DBTable)
	}

	return nil
}

// plateWorker method is initialized at startup to grab plate requests from a channel then process the request (stdout or db save)
func (s *Server) plateWorker() {
	var start time.Time
	var records int
	for {
		select {
		case pr := <-s.queuePlates:
			s.processPlate(records, &pr)
			records++
			// start a timer
			if records == 1 {
				start = time.Now()
			}
		default:
			// after queue is empty, reset the record count and show total processed and duration
			if records > 0 {
				log.Printf("Processed %v plates in %v\n", records, time.Since(start))
				records = 0
			}
		}
	}
}

// processPlate method accepts a index integer and grpc request message for processing.
// Will output requests received to STDOUT or save to the database table.
func (s *Server) processPlate(i int, pr *pb.PlateRequest) {
	switch s.OutputType {
	case "db":
		err := s.dbInsert(pr)
		if err != nil {
			log.Printf("DB Error (%+v) for item (%v) -- %v", err, i, pr)
		}
	case "stdout":
		log.Printf("Stdout (%v) -- %+v", i, pr)
	}
}

// dbConfig method initializes the database connection and checks if the table exists or will create it immediately.
func (s *Server) dbConfig() error {
	if s.dbInit {
		return nil
	}

	var err error
	params := fmt.Sprintf("host=%v user=%v password=%v sslmode=disable connect_timeout=5", s.DBHost, s.DBUsername, s.DBPassword)
	s.dbConn, err = sqlx.Connect("postgres", params)
	if err != nil {
		return err
	}

	// setup DB open/idle connections for better DB connection availability
	s.dbConn.SetMaxOpenConns(25)
	s.dbConn.SetMaxIdleConns(25)
	s.dbConn.SetConnMaxLifetime(5 * time.Minute)

	// check if the DB table exists or create it
	tableExist := false
	rows, err := s.dbConn.Queryx(fmt.Sprintf("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '%v')", s.DBTable))
	if err != nil {
		return err
	}
	// check query response
	for rows.Next() {
		m := map[string]interface{}{}
		rows.MapScan(m)
		if val, ok := m["exists"]; ok && val == true {
			tableExist = true
		}
	}

	// exists or create table
	if !tableExist {
		s.dbConn.MustExec(dbSchema)
		log.Println("DB Init - Table created")
	} else {
		log.Println("DB Init - Table exists")
	}

	// DB table is ready
	s.dbInit = true

	return nil
}

// SetupDB method calls dbConfig method from outside package for DB setup initiliazation
func (s *Server) SetupDB() error {
	return s.dbConfig()
}

// dbInsert method accepts a gRPC request (plate request) then inserts the record to the DB table
func (s *Server) dbInsert(pr *pb.PlateRequest) error {
	// check if DB table is ready for inserts
	if s.dbInit == false {
		return errors.New("DB not ready")
	}

	// insert enrty to table
	_, err := s.dbConn.NamedExec(`INSERT INTO plates (plate, well, runtime, optical_density, run, corrected_optical_density) VALUES (:plate, :well, :runtime, :optical_density, :run, :corrected_optical_density)`,
		map[string]interface{}{
			"plate":                     pr.Plate,
			"well":                      pr.Well,
			"runtime":                   pr.Runtime,
			"optical_density":           pr.OpticalDensity,
			"run":                       pr.Run,
			"corrected_optical_density": pr.CorrectedOpticalDensity,
		})

	if err != nil {
		return err
	}

	return nil
}
