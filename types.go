package medego

import (
	"database/sql"
	"log"
	"net/url"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/PrinMeshia/medego/cache"
	"github.com/PrinMeshia/medego/mailer"
	"github.com/PrinMeshia/medego/render"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/robfig/cron/v3"
)

type Medego struct {
	AppName       string
	Debug         bool
	Version       string
	ErrorLog      *log.Logger
	InfoLog       *log.Logger
	RootPath      string
	Routes        *chi.Mux
	Render        *render.Render
	Session       *scs.SessionManager
	DB            Database
	JetViews      *jet.Set
	Config        ServerConfig
	EncryptionKey string
	Cache         cache.Cache
	Scheduler     *cron.Cron
	Mail          mailer.Mail
	Server        Server
}
type Server struct {
	ServerName string
	Port       string
	Secure     bool
	URL        string
}
type ServerConfig struct {
	Port         string
	Renderer     string
	Cookie       cookieConfig
	SessionType  string
	IdleTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Database     databaseConfig
	redis        redisConfig
}

type initPaths struct {
	rootPath    string
	folderNames []string
}

type cookieConfig struct {
	name     string
	lifetime string
	persist  string
	secure   string
	domain   string
}

type databaseConfig struct {
	dsn      string
	database string
}

type Database struct {
	DataType string
	Pool     *sql.DB
}

type Validation struct {
	Data   url.Values
	Errors map[string]string
}

type Encryption struct {
	Key []byte
}

type redisConfig struct {
	host     string
	password string
	prefix   string
}
