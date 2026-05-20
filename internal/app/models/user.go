package models

import (
	"gorm.io/gorm"
	"k8soperation/internal/errorcode"
	"strconv"
	"time"
)

type User struct {
	Username string `json:"username" description:"用户名"`
	Password string `json:"-" description:"密码"`
	*Base
}

func NewUser() *User {
	return &User{}
}

func (u *User) TableName() string {
	return "user"
}

func (u *User) Create(db *gorm.DB) error {
	return db.Create(&u).Error
}

func (u *User) Delete(db *gorm.DB) error {
	var user User
	if err := db.Where("id=? and is_del=?", u.ID, 0).First(&user).Error; err != nil {
		return err
	}

	nowTime := uint32(time.Now().Unix())
	user.IsDel = 1
	user.DeletedAt = nowTime
	user.ModifiedAt = nowTime

	if err := db.Updates(&user).Error; err != nil {
		return err
	}
	return nil
}

func (u *User) Update(db *gorm.DB, values interface{}) error {
	tx := db.Model(u).
		Where("id=? AND is_del=?", u.ID, 0).
		Updates(values)

	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errorcode.ErrorUserUpdateFail
	}
	return nil
}

func (u *User) List(db *gorm.DB, page, limit int) ([]*User, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 1000 {
		limit = 20
	}

	var users []*User
	q := db.Model(&User{}).Where("is_del = 0")
	if u.Username != "" {
		q = q.Where("username = ?", u.Username)
	}

	offset := (page - 1) * limit
	if err := q.Order("id DESC").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (u *User) GetByName(db *gorm.DB) (*User, error) {
	var user *User
	if u.Username != "" {
		db.Where("username = ?", u.Username).First(&user)
	}
	return user, nil
}

func (u *User) GetUserByID(db *gorm.DB, id string) User {
	var user = NewUser()
	db.Where("id", id).First(&user)
	return *user
}

func (u *User) GetStringID() string {
	return strconv.FormatUint(uint64(u.ID), 10)
}
