package generator

import (
	"fmt"
	"github.com/fishdivinity/BeeCount-Cloud/services/config/internal/model"
)

// GenerateDatabaseConfig 生成数据库配置内容
func GenerateDatabaseConfig(cfg *model.DatabaseConfig) string {
	return `# Database Configuration
database:
  active: ` + cfg.Active + ` # sqlite, mysql, postgres
  sqlite:
    path: ` + cfg.SQLite.Path + ` # SQLite database file path
  mysql:
    host: ` + cfg.MySQL.Host + ` # MySQL host address
    port: ` + fmt.Sprintf("%d", cfg.MySQL.Port) + ` # MySQL port
    username: ` + cfg.MySQL.Username + ` # MySQL username
    password: ` + cfg.MySQL.Password + ` # MySQL password
    database: ` + cfg.MySQL.Database + ` # MySQL database name
    charset: ` + cfg.MySQL.Charset + ` # MySQL charset
    parse_time: ` + fmt.Sprintf("%t", cfg.MySQL.ParseTime) + ` # Whether to parse time
    loc: ` + cfg.MySQL.Loc + ` # Timezone
  postgres:
    host: ` + cfg.Postgres.Host + ` # PostgreSQL host address
    port: ` + fmt.Sprintf("%d", cfg.Postgres.Port) + ` # PostgreSQL port
    username: ` + cfg.Postgres.Username + ` # PostgreSQL username
    password: ` + cfg.Postgres.Password + ` # PostgreSQL password
    database: ` + cfg.Postgres.Database + ` # PostgreSQL database name
    sslmode: ` + cfg.Postgres.SSLMode + ` # SSL mode
    timezone: ` + cfg.Postgres.Timezone + ` # Timezone
  pool:
    max_idle_conns: ` + fmt.Sprintf("%d", cfg.Pool.MaxIdleConns) + ` # Maximum number of idle connections
    max_open_conns: ` + fmt.Sprintf("%d", cfg.Pool.MaxOpenConns) + ` # Maximum number of open connections
    conn_max_lifetime: ` + fmt.Sprintf("%v", cfg.Pool.ConnMaxLifetime) + ` # Maximum connection lifetime
    conn_max_idle_time: ` + fmt.Sprintf("%v", cfg.Pool.ConnMaxIdleTime) + ` # Maximum idle connection time
`
}
