package medego

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/PrinMeshia/medego/cache"
	"github.com/PrinMeshia/medego/mailer"
	"github.com/PrinMeshia/medego/render"
	"github.com/PrinMeshia/medego/session"
	"github.com/dgraph-io/badger/v4"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

const version = "1.0.0"

var redisCache *cache.RedisCache
var badgerCache *cache.BadgerCache
var redisPool *redis.Pool
var badgerConn *badger.DB

const (
	envFileName  = ".env"
	rootPathName = ""
)
const (
	defaultPort         = "8080"
	defaultRenderer     = "html"
	defaultIdleTimeout  = 30 * time.Second
	defaultReadTimeout  = 30 * time.Second
	defaultWriteTimeout = 600 * time.Second
)

func (c *Medego) New(rootPath string) error {
	pathConfig := initPaths{
		rootPath: rootPath,
		folderNames: []string{
			"src/middleware", "src/handlers", "src/data",
			"migrations", "templates", "public", "mail",
			"tmp/logs", "tmp/cache"},
	}
	if err := c.Init(pathConfig); err != nil {
		return err
	}

	if err := c.checkDotEnv(rootPath); err != nil {
		return err
	}

	if err := godotenv.Load(rootPath + "/" + envFileName); err != nil {
		return err
	}
	//create logger
	infoLog, errorLog := c.startLoggers()

	//database connection
	if os.Getenv("DATABASE_TYPE") != "" {
		db, err := c.OpenDB(os.Getenv("DATABASE_TYPE"), c.BuildDSN())
		if err != nil {
			errorLog.Println(err)
			os.Exit(1)
		}
		c.DB = Database{
			DataType: os.Getenv("DATABASE_TYPE"),
			Pool:     db,
		}
	}

	scheduler := cron.New()
	c.Scheduler = scheduler

	if os.Getenv("CACHE") == "redis" || os.Getenv("SESSION_TYPE") == "redis" {
		redisCache = c.createClientRedisCache()
		c.Cache = redisCache
		redisPool = redisCache.Conn
	}

	if os.Getenv("CACHE") == "badger" || os.Getenv("SESSION_TYPE") == "badger" {
		badgerCache = c.createClientBadgerCache()
		c.Cache = badgerCache
		badgerConn = badgerCache.Conn

		if _, err := c.Scheduler.AddFunc("@daily", func() {
			_ = badgerCache.Conn.RunValueLogGC(0.7)
		}); err != nil {
			return err
		}
	}

	c.InfoLog = infoLog
	c.ErrorLog = errorLog

	c.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	c.Version = version
	c.RootPath = rootPath
	c.Mail = c.createMailer()
	c.Routes = c.routes().(*chi.Mux)

	c.Config = ServerConfig{
		Port:     os.Getenv("PORT"),
		Renderer: os.Getenv("RENDERER"),
		Cookie: cookieConfig{
			name:     os.Getenv("COOKIE_NAME"),
			lifetime: os.Getenv("COOKIE_LIFETIME"),
			persist:  os.Getenv("COOKIE_PERSISTS"),
			secure:   os.Getenv("COOKIE_SECURE"),
			domain:   os.Getenv("COOKIE_DOMAIN"),
		},
		SessionType: os.Getenv("SESSION_TYPE"),
		Database: databaseConfig{
			database: os.Getenv("DATABASE_TYPE"),
			dsn:      c.BuildDSN(),
		},
		redis: redisConfig{
			host:     os.Getenv("REDIS_HOST"),
			password: os.Getenv("REDIS_PASSWORD"),
			prefix:   os.Getenv("REDIS_PREFIX"),
		},
	}
	secure := true
	if strings.ToLower(os.Getenv("SECURE")) == "false" {
		secure = false
	}
	c.Server = Server{
		ServerName: os.Getenv("SERVER_NAME"),
		Port:       os.Getenv("PORT"),
		Secure:     secure,
		URL:        os.Getenv("APP_URL"),
	}

	sess := session.Session{
		CookieLifetime: c.Config.Cookie.lifetime,
		CookiePersist:  c.Config.Cookie.persist,
		CookieName:     c.Config.Cookie.name,
		SessionType:    c.Config.SessionType,
		CookieDomain:   c.Config.Cookie.domain,
	}

	switch c.Config.SessionType {
	case "redis":
		sess.RedisPool = redisCache.Conn
	case "mysql", "postgres", "mariadb", "postgresql":
		sess.DBPool = c.DB.Pool
	}

	c.Session = sess.InitSession()
	c.EncryptionKey = os.Getenv("KEY")
	if c.Config.Port == "" {
		c.Config.Port = defaultPort
	}

	if c.Config.Renderer == "" {
		c.Config.Renderer = defaultRenderer
	}
	if c.Debug {
		var views = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/templates", rootPath)),
			jet.InDevelopmentMode(),
		)
		c.JetViews = views

	} else {
		var views = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/templates", rootPath)),
		)
		c.JetViews = views
	}

	c.createRenderer()
	go c.Mail.ListenForMail()
	return nil
}
func (c *Medego) Init(p initPaths) error {
	root := p.rootPath
	for _, path := range p.folderNames {
		// Créez un chemin complet en utilisant la concaténation de chaînes
		fullPath := fmt.Sprintf("%s/%s", root, path)

		// Créez le dossier s'il n'existe pas
		err := c.CreateDirIfNotExists(fullPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Medego) ListenAndServe() {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", c.Config.Port),
		ErrorLog:     c.ErrorLog,
		Handler:      c.Routes,
		IdleTimeout:  defaultIdleTimeout,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
	}

	if c.DB.Pool != nil {
		defer c.DB.Pool.Close()
	}

	if redisPool != nil {
		defer redisPool.Close()
	}
	if badgerConn != nil {
		defer badgerConn.Close()
	}

	c.InfoLog.Printf("Listening on port %s", c.Config.Port)
	c.ErrorLog.Fatal(srv.ListenAndServe())
}

func (c *Medego) checkDotEnv(path string) error {
	return c.CreateFileIfNotExists(fmt.Sprintf("%s/%s/%s", path, rootPathName, envFileName))
}

func (c *Medego) startLoggers() (*log.Logger, *log.Logger) {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	return infoLog, errorLog
}

func (c *Medego) createRenderer() {
	myRenderer := render.Render{
		Renderer: c.Config.Renderer,
		RootPath: c.RootPath,
		Port:     c.Config.Port,
		JetViews: c.JetViews,
		Session:  c.Session,
	}
	c.Render = &myRenderer

}

func (c *Medego) createMailer() mailer.Mail {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	m := mailer.Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Templates:   c.RootPath + "/mail",
		Host:        os.Getenv("SMTP_HOST"),
		Port:        port,
		Username:    os.Getenv("SMTP_USERNAME"),
		Password:    os.Getenv("SMTP_PASSWORD"),
		Encryption:  os.Getenv("SMTP_ENCRYPTION"),
		FromName:    os.Getenv("FROM_NAME"),
		FromAddress: os.Getenv("FROM_ADDRESS"),
		Jobs:        make(chan mailer.Message, 20),
		Results:     make(chan mailer.Result, 20),
		API:         os.Getenv("MAILER_API"),
		APIKey:      os.Getenv("MAILER_KEY"),
		APIUrl:      os.Getenv("MAILER_URL"),
	}
	return m
}

func (c *Medego) createClientRedisCache() *cache.RedisCache {
	cacheClient := cache.RedisCache{
		Conn:   c.createRedisPool(),
		Prefix: c.Config.redis.prefix,
	}
	return &cacheClient
}

func (c *Medego) createClientBadgerCache() *cache.BadgerCache {
	cacheClient := cache.BadgerCache{
		Conn: c.createBadgerConn(),
	}
	return &cacheClient
}

func (c *Medego) createRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,
		MaxActive:   10000,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp",
				c.Config.redis.host,
				redis.DialPassword(c.Config.redis.password))
		},
		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			_, err := conn.Do("PING")
			return err
		},
	}
}
func (c *Medego) createBadgerConn() *badger.DB {
	db, err := badger.Open(badger.DefaultOptions(c.RootPath + "/tmp/badger"))
	if err != nil {
		return nil
	}
	return db
}
func (c *Medego) BuildDSN() string {
	dsn := ""

	switch os.Getenv("DATABASE_TYPE") {
	case "postgres", "postgresql":
		dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_PORT"),
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_NAME"),
			os.Getenv("DATABASE_SSL_MODE"))

		if os.Getenv("DATABASE_PASS") != "" {
			dsn = fmt.Sprintf("%s password=%s", dsn, os.Getenv("DATABASE_PASS"))
		}
	default:

	}

	return dsn
}
