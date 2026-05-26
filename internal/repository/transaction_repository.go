// internal/repository/transaction.go
package repository

import "gorm.io/gorm"

// DBInterface - интерфейс для операций с БД
type DBInterface interface {
	First(dest interface{}, conds ...interface{}) *gorm.DB
	Find(dest interface{}, conds ...interface{}) *gorm.DB
	Create(value interface{}) *gorm.DB
	Save(value interface{}) *gorm.DB
	Delete(value interface{}, conds ...interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) *gorm.DB
	Preload(query string, args ...interface{}) *gorm.DB
	Model(value interface{}) *gorm.DB
	Exec(sql string, values ...interface{}) *gorm.DB
	Raw(sql string, values ...interface{}) *gorm.DB
	Scan(dest interface{}) *gorm.DB
	Count(count *int64) *gorm.DB
	Begin() *gorm.DB
	Commit() *gorm.DB
	Rollback() *gorm.DB
	Error() error
}

// GormDB адаптер для gorm.DB
type GormDB struct {
	db *gorm.DB
}

func NewGormDB(db *gorm.DB) *GormDB {
	return &GormDB{db: db}
}

func (g *GormDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	return g.db.First(dest, conds...)
}

func (g *GormDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	return g.db.Find(dest, conds...)
}

func (g *GormDB) Create(value interface{}) *gorm.DB {
	return g.db.Create(value)
}

func (g *GormDB) Save(value interface{}) *gorm.DB {
	return g.db.Save(value)
}

func (g *GormDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	return g.db.Delete(value, conds...)
}

func (g *GormDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	return g.db.Where(query, args...)
}

func (g *GormDB) Preload(query string, args ...interface{}) *gorm.DB {
	return g.db.Preload(query, args...)
}

func (g *GormDB) Model(value interface{}) *gorm.DB {
	return g.db.Model(value)
}

func (g *GormDB) Exec(sql string, values ...interface{}) *gorm.DB {
	return g.db.Exec(sql, values...)
}

func (g *GormDB) Raw(sql string, values ...interface{}) *gorm.DB {
	return g.db.Raw(sql, values...)
}

func (g *GormDB) Scan(dest interface{}) *gorm.DB {
	return g.db.Scan(dest)
}

func (g *GormDB) Count(count *int64) *gorm.DB {
	return g.db.Count(count)
}

func (g *GormDB) Begin() *gorm.DB {
	return g.db.Begin()
}

func (g *GormDB) Commit() *gorm.DB {
	return g.db.Commit()
}

func (g *GormDB) Rollback() *gorm.DB {
	return g.db.Rollback()
}

func (g *GormDB) Error() error {
	return g.db.Error
}
