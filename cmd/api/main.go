package main

import (
	"flag"
	"fmt" 
	"log" 
	"net/http"
	"os" 
	"time"
	"context"
	"database/sql"

	// Import the pq driver so that it can register itself with the database/sql
	// package. Note that we alias this import to the blank identifier, to stop the Go
	// compiler complaining that the package isn't being used.
	_ "github.com/lib/pq"
	"greenlight.nursultandias.net/internal/data"
)

// application version number. 
// Later it will generate this automatically at build time
// but for now it is hard-coded
const version = "1.0.0"

// The config structure holds port number(port) and stage(env) of the application.
// Add a db struct field to hold the configuration settings for our database connection
// pool. It holds the DSN, which we will read in from a command-line flag.
// maxOpenConns, maxIdleConns and maxIdleTime fields to hold the configuration
// settings for the connection pool.
type config struct {
	port	int
	env		string
	db		struct {
		dsn				string
		maxOpenConns	int
		maxIdleConns	int
		maxIdleTime		string
	}
}

// the application structure holds top config structure and logger. 
// logger is usefull since it does: 
// ⭐ If something goes wrong, the logger helps you understand what happened by recording errors, 
// ⭐ stack traces, and the sequence of events that led to the problem
// ⭐ Track important events like when the API starts/stops
// ⭐ Monitor performance (how long requests take)
// ⭐ See which endpoints are being called most frequently
// ⭐ Keep track of important operations (like database changes)
// ⭐ Record who did what and when
// Add a models field to hold our new Models struct.
type application struct {
	config	config
	logger	*log.Logger
	models data.Models
}

func main() {
	var cfg config

	//(struct pointer, flag name, default value, description)
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// Read the DSN value from the db-dsn command-line flag into the config struct.
	// Use the value of the GREENLIGHT_DB_DSN environment variable as the default value
	// for our db-dsn command-line flag.
	flag.StringVar(&cfg.db.dsn, "db-dsn",  os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")

	// Read the connection pool settings from command-line flags into the config struct.
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	flag.Parse()

	// Initialize a new logger which writes messages to the standard out stream, 
	// prefixed with the current date and time.
	logger := log.New(os.Stdout, "", log.Ldate | log.Ltime)

	// Call the openDB() helper function (see below after main function) to create the connection pool,
	// passing in the config struct. If this returns an error, we log it and exit the
	// application immediately.
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	// Defer a call to db.Close() so that the connection pool is closed before the
	// main() function exits.
	defer db.Close()

	// Also log a message to say that the connection pool has been successfully
	// established.
	logger.Printf("database connection pool established")

	// Use the data.NewModels() function to initialize a Models struct, passing in the
	// connection pool as a parameter.
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	// Declare a HTTP server with some sensible timeout settings, which listens on the 
	// port provided in the config struct and uses the servemux we created above as the 
	// handler
	srv := &http.Server{
		Addr:	fmt.Sprintf(":%d", cfg.port),
		Handler: app.routes(),
		IdleTimeout: time.Minute,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)
}


// The openDB() function returns a sql.DB connection pool.
func openDB(cfg config) (*sql.DB, error) {
	// Use sql.Open() to create an empty connection pool, using the DSN from the config struct.
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Set the maximum number of open (in-use + idle) connections in the pool.
	// Note that passing a value less than or equal to 0 will mean there is no limit.
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	// Set the maximum number of idle connections in the pool. Again, passing a value
	// less than or equal to 0 will mean there is no limit.
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	// Use the time.ParseDuration() function to convert the idle timeout duration string
	// to a time.Duration type.
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	// Set the maximum idle timeout.
	db.SetConnMaxIdleTime(duration)

	// Create a context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use PingContext() to establish a new connection to the database, passing in the
	// context we created above as a parameter. If the connection couldn't be
	// established successfully within the 5 second deadline, then this will return an error.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	// Return the sql.DB connection pool.
	return db, nil
}