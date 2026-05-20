package migration

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"k8soperation/internal/app/models"

	"gorm.io/gorm"
)

type Handler func(db *gorm.DB) error

type entry struct {
	version string
	name    string
	handler Handler
}

type Registry struct {
	mutex      sync.Mutex
	migrations map[string]entry
}

var Default = NewRegistry()

func NewRegistry() *Registry {
	return &Registry{migrations: make(map[string]entry)}
}

func Register(version string, handler Handler) {
	Default.Register(version, "", handler)
}

func RegisterNamed(version, name string, handler Handler) {
	Default.Register(version, name, handler)
}

func Run(db *gorm.DB) error {
	return Default.Run(db)
}

func (r *Registry) Register(version, name string, handler Handler) {
	version = strings.TrimSpace(version)
	if version == "" {
		panic("migration version is empty")
	}
	if handler == nil {
		panic(fmt.Sprintf("migration %s handler is nil", version))
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.migrations[version]; ok {
		panic(fmt.Sprintf("migration %s already registered", version))
	}
	r.migrations[version] = entry{version: version, name: name, handler: handler}
}

func (r *Registry) Run(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("migration db is nil")
	}
	if err := db.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").AutoMigrate(&models.Migration{}); err != nil {
		return fmt.Errorf("create migration table failed: %w", err)
	}

	for _, item := range r.sorted() {
		applied, err := isApplied(db, item.version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if err := runOne(db, item); err != nil {
			return err
		}
	}
	return nil
}

func (r *Registry) sorted() []entry {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	versions := make([]string, 0, len(r.migrations))
	for version := range r.migrations {
		versions = append(versions, version)
	}
	sort.Strings(versions)

	items := make([]entry, 0, len(versions))
	for _, version := range versions {
		items = append(items, r.migrations[version])
	}
	return items
}

func isApplied(db *gorm.DB, version string) (bool, error) {
	var count int64
	if err := db.Model(&models.Migration{}).Where("version = ?", version).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check migration %s failed: %w", version, err)
	}
	return count > 0, nil
}

func runOne(db *gorm.DB, item entry) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := item.handler(tx); err != nil {
			return fmt.Errorf("run migration %s failed: %w", item.version, err)
		}
		name := item.name
		if name == "" {
			name = item.version
		}
		record := models.Migration{
			Version:   item.version,
			Name:      name,
			CreatedAt: uint32(time.Now().Unix()),
		}
		if err := tx.Create(&record).Error; err != nil {
			return fmt.Errorf("record migration %s failed: %w", item.version, err)
		}
		return nil
	})
}

func VersionFromFilename(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}
