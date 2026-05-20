package migrate

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"k8soperation/cmd/migrate/migration"
	_ "k8soperation/cmd/migrate/migration/version"
	"k8soperation/pkg/app"
	"k8soperation/pkg/database"
	applogger "k8soperation/pkg/logger"
	"k8soperation/pkg/setting"
	"k8soperation/pkg/setting/types"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	gormlogger "gorm.io/gorm/logger"
)

var migrationTemplate = template.Must(template.New("migration").Parse(`package {{.Package}}

import (
	"gorm.io/gorm"
	"runtime"
	"k8s-manager/cmd/migrate/migration"

)

func init() {
	_, fileName, _, _ := runtime.Caller(0)
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _{{.GenerateTime}}Test)
}

func _{{.GenerateTime}}Test(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var err error
		
	    // TODO: 这里开始写入要变更的内容

	    // TODO: 例如 修改表字段 使用过程中请删除此段代码
		//err := tx.Migrator().RenameColumn(&models.SysConfig{}, "config_id", "id")
		//if err != nil {
		// 	return err
		//}

		// TODO: 例如 新增表结构 使用过程中请删除此段代码
		//err = tx.Debug().Migrator().AutoMigrate(
        //		new(models.CasbinRule),
        // 		)
        //if err != nil {
        // 	return err
        //}


		return tx.Create(&pkgmodels.Migration{
			Version: version,
		}).Error
	})
}
`))

type templateData struct {
	Version    string
	Name       string
	ConstName  string
	FuncSuffix string
}

func NewCommand(configFile *string) *cobra.Command {
	var generateName string

	cmd := &cobra.Command{
		Use:          "migrate",
		Aliases:      []string{"m"},
		Short:        "Run database migrations",
		Example:      "k8s-manager migrate -c configs/config.yaml",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if generateName != "" {
				path, err := generateFile(generateName)
				if err != nil {
					return err
				}
				cmd.Printf("generated migration: %s\n", path)
				return nil
			}
			return run(configValue(configFile))
		},
	}
	cmd.Flags().StringVarP(&generateName, "generate", "g", "", "Generate migration file with the provided name")
	return cmd
}

func configValue(configFile *string) string {
	if configFile == nil {
		return ""
	}
	return *configFile
}

func run(configFile string) error {
	a, err := setupApp(configFile)
	if err != nil {
		return err
	}
	defer func() {
		_ = a.Logger.Sync()
		_ = a.BizLogger.Sync()
		if a.SQLDB != nil {
			_ = a.SQLDB.Close()
		}
	}()

	a.Logger.Info("migrate database start")
	if err := migration.Run(a.DB); err != nil {
		a.Logger.Error("migrate database failed", zap.Error(err))
		return err
	}
	a.Logger.Info("migrate database success")
	return nil
}

func setupApp(configFile string) (*app.App, error) {
	config := types.AppConfig{}
	if err := setting.LoadAppConfig(&config, configFile); err != nil {
		return nil, err
	}

	a := app.NewApp()
	applySetting(a, &config)
	if err := setupLogger(a); err != nil {
		return nil, err
	}
	if err := setupDB(a); err != nil {
		return nil, err
	}
	return a, nil
}

func applySetting(a *app.App, config *types.AppConfig) {
	a.AppConfig = config
	if config == nil {
		return
	}
	a.ServerSetting = config.ServerSetting
	a.AppSetting = config.AppSetting
	a.DatabaseSetting = config.DatabaseSetting
	a.CacheSetting = config.CacheSetting
	a.PodLogSetting = config.PodLogSetting
	a.NodeSetting = config.NodeSetting
}

func setupLogger(a *app.App) error {
	if a.AppSetting == nil {
		return fmt.Errorf("AppSetting is nil")
	}
	if err := ensureDir(a.AppSetting.LogFileName); err != nil {
		return err
	}
	if a.AppSetting.BusinessLogFileName == "" {
		a.AppSetting.BusinessLogFileName = "logs/app.log"
	}
	if err := ensureDir(a.AppSetting.BusinessLogFileName); err != nil {
		return err
	}

	level := applogger.WithLevel(a.AppSetting.LogLevel)
	a.Logger = applogger.NewLogger(level, applogger.RotateOptions{
		FileName:   a.AppSetting.LogFileName,
		MaxSize:    a.AppSetting.LogMaxSize,
		MaxBackups: a.AppSetting.LogMaxBackup,
		MaxAge:     a.AppSetting.LogMaxAge,
		Compress:   a.AppSetting.LogCompress,
	}, applogger.AddCaller(), applogger.AddCallerSkip(1), applogger.AddStacktrace(applogger.ErrorLevel))
	a.BizLogger = applogger.NewBusinessLogger(applogger.RotateOptions{
		FileName:   a.AppSetting.BusinessLogFileName,
		MaxSize:    a.AppSetting.LogMaxSize,
		MaxBackups: a.AppSetting.LogMaxBackup,
		MaxAge:     a.AppSetting.LogMaxAge,
		Compress:   a.AppSetting.LogCompress,
	}, applogger.AddCaller(), applogger.AddCallerSkip(1))
	return nil
}

func ensureDir(filePath string) error {
	dir := filepath.Dir(filePath)
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("create log dir %q: %w", dir, err)
			}
			return nil
		}
		return fmt.Errorf("stat log dir %q: %w", dir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("log path %q exists but is not a directory", dir)
	}
	return nil
}

func setupDB(a *app.App) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%t&loc=Local&timeout=1s&readTimeout=2s&writeTimeout=2s",
		a.DatabaseSetting.Username,
		a.DatabaseSetting.Password,
		a.DatabaseSetting.Host,
		a.DatabaseSetting.Port,
		a.DatabaseSetting.DBName,
		a.DatabaseSetting.Charset,
		a.DatabaseSetting.ParseTime,
	)

	var err error
	a.DB, a.SQLDB, err = database.Connect(mysql.New(mysql.Config{DSN: dsn}), gormlogger.Default.LogMode(gormlogger.Info))
	if err != nil {
		return fmt.Errorf("connect db failed: %w", err)
	}

	a.SQLDB.SetMaxOpenConns(a.DatabaseSetting.MaxOpenConns)
	a.SQLDB.SetMaxIdleConns(a.DatabaseSetting.MaxIdleConns)
	a.SQLDB.SetConnMaxLifetime(time.Duration(a.DatabaseSetting.MaxLifeSeconds) * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := a.SQLDB.PingContext(ctx); err != nil {
		return fmt.Errorf("db ping failed: %w", err)
	}
	return nil
}

func generateFile(name string) (string, error) {
	name = normalizeName(name)
	if name == "" {
		return "", fmt.Errorf("migration name is empty")
	}

	version := time.Now().Format("20060102150405")
	data := templateData{
		Version:    version,
		Name:       name,
		ConstName:  "migration" + version,
		FuncSuffix: version,
	}

	var buf bytes.Buffer
	if err := migrationTemplate.Execute(&buf, data); err != nil {
		return "", err
	}

	dir := filepath.Join("cmd", "migrate", "migration", "version")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create migration version dir failed: %w", err)
	}
	path := filepath.Join(dir, fmt.Sprintf("%s_%s.go", version, name))
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return "", fmt.Errorf("write migration file failed: %w", err)
	}
	return path, nil
}

func normalizeName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(name, "_")
	return strings.Trim(name, "_")
}
