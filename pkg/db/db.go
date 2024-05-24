package db

import (
	// "context"
	// "fmt"
	"fmt"
	"strconv"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/dolphindb/api-go/api"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

type DBConfig struct {
	Autologin    bool
	Password     string
	Python       bool
	URL          string
	Username     string
	Verbose      bool
	PoolCapacity string
}

type DataSource struct {
	Conn   *api.DBConnectionPool
	Config DBConfig
}

var (
	dataSourceMap     = make(map[string]*DataSource)
	dataSourceMapLock sync.RWMutex
)

// getDatasource returns a connection for the given UUID and DBConfig.
func GetDatasource(uuid string, config DBConfig) (*api.DBConnectionPool, error) {
	dataSourceMapLock.Lock()
	defer dataSourceMapLock.Unlock()

	// Check if the datasource already exists
	if ds, exists := dataSourceMap[uuid]; exists {
		if ds.Config == config {
			// Config hasn't changed, return the existing connection
			log.DefaultLogger.Info("Reused Connection")
			return ds.Conn, nil
		} else {
			// Config has changed, close the old connection and create a new one
			ds.Conn.Close()
			delete(dataSourceMap, uuid)
		}
	}

	// Create a new connection
	log.DefaultLogger.Info("Get Connection Pool")
	log.DefaultLogger.Info(spew.Sdump(config))

	capacity, err := strconv.Atoi(config.PoolCapacity)
	if err != nil {
		return nil, fmt.Errorf("pool capacity convert to num error %v", err)
	}

	poolOpt := &api.PoolOption{
		Address:  config.URL,
		UserID:   config.Username,
		Password: config.Password,
		PoolSize: capacity,
	}
	pool, err := api.NewDBConnectionPool(poolOpt)

	if err != nil {
		return nil, err
	}
	log.DefaultLogger.Debug(fmt.Sprintf("Connected to DB connection pool with capacity %d", capacity))
	// Store the new datasource
	dataSourceMap[uuid] = &DataSource{
		Conn:   pool,
		Config: config,
	}

	return pool, nil
}

// Example of how to use getDatasource
// func LoadTableWithDatasource(uuid, dbPath, tbName string, config DBConfig) (interface{}, error) {
// 	conn, err := GetDatasource(uuid, config)
// 	if err != nil {
// 		return nil, err
// 	}

// 	tb, err := conn.RunScript(fmt.Sprintf("select * from loadTable('%s','%s')", dbPath, tbName))
// 	if err != nil {
// 		return nil, err
// 	}

// 	return tb, nil
// }
