package initialize

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/thedevsaddam/govalidator"
	"k8soperation/pkg/app"
)

func SetupValidator(a *app.App) error {
	govalidator.AddCustomRule("not_exists", func(field, rule, message string, value interface{}) error {
		if !strings.HasPrefix(rule, "not_exists:") {
			return errors.New("not_exists 规则格式错误")
		}

		raw := strings.TrimSpace(strings.TrimPrefix(rule, "not_exists:"))
		parts := strings.Split(raw, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
			return errors.New("not_exists 规则参数错误，期望 not_exists:表名,字段名[,except_id=ID]")
		}
		tableName := parts[0]
		dbField := parts[1]

		if !isAllowedTable(tableName) || !isAllowedColumn(tableName, dbField) {
			return errors.New("非法的表名或字段名")
		}

		val := strings.TrimSpace(fmt.Sprint(value))
		if val == "" {
			return nil
		}

		var exceptIDStr string
		if len(parts) >= 3 && parts[2] != "" {
			if strings.Contains(parts[2], "=") {
				kv := strings.SplitN(parts[2], "=", 2)
				if len(kv) == 2 && strings.TrimSpace(kv[0]) == "except_id" {
					exceptIDStr = strings.TrimSpace(kv[1])
				}
			} else {
				exceptIDStr = parts[2]
			}
		}

		var exceptID any
		if exceptIDStr != "" {
			if id64, err := strconv.ParseInt(exceptIDStr, 10, 64); err == nil {
				exceptID = id64
			} else {
				return errors.New("except_id 必须为数字")
			}
		}

		q := a.DB.Table(tableName).Where(fmt.Sprintf("%s = ?", dbField), val)
		if exceptID != nil {
			q = q.Where("id <> ?", exceptID)
		}

		var count int64
		if err := q.Count(&count).Error; err != nil {
			return errors.New("系统繁忙，请稍后再试")
		}
		if count > 0 {
			if message != "" {
				return errors.New(message)
			}
			return fmt.Errorf("%s 已存在", field)
		}
		return nil
	})

	return nil
}
