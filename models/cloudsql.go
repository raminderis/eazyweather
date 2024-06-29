package models

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"cloud.google.com/go/cloudsqlconn"
	mssql "github.com/denisenkom/go-mssqldb"
)

func DefaultCloudSqlConfig() CloudSqlConfig {
	return CloudSqlConfig{
		User:                   "pgeazy",
		Password:               "pgeazypassword",
		DBname:                 "eazyweather",
		InstanceConnectionName: "micro-liberty-413905:us-west1:sqlinstance",
	}
}

type CloudSqlConfig struct {
	User                   string
	Password               string
	DBname                 string
	InstanceConnectionName string
}

func (cfg CloudSqlConfig) String() string {
	return fmt.Sprintf("user id=%s;password=%s;database=%s;", cfg.User, cfg.Password, cfg.DBname)
}

type csqlDialer struct {
	dialer     *cloudsqlconn.Dialer
	connName   string
	usePrivate bool
}

// DialContext adheres to the mssql.Dialer interface.
func (c *csqlDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	var opts []cloudsqlconn.DialOption
	if c.usePrivate {
		opts = append(opts, cloudsqlconn.WithPrivateIP())
	}
	return c.dialer.Dial(ctx, c.connName, opts...)
}

func ConnectWithConnector(config CloudSqlConfig) (*sql.DB, error) {
	dbURI := config.String()
	c, err := mssql.NewConnector(dbURI)
	if err != nil {
		return nil, fmt.Errorf("cloudsqlconn.NewDailer: %w", err)
	}
	dialer, err := cloudsqlconn.NewDialer(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cloudsqlconn.NewDailer: %w", err)
	}
	c.Dialer = &csqlDialer{
		dialer:     dialer,
		connName:   config.InstanceConnectionName,
		usePrivate: false,
	}
	db := sql.OpenDB(c)
	return db, nil
}
