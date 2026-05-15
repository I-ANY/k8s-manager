package models

import (
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"k8soperation/global"
	"k8soperation/internal/errorcode"
)

// K8sCluster K8s集群结构体
type K8sCluster struct {
	ClusterName    string `json:"cluster_name" description:"集群名"`
	ClusterVersion string `json:"cluster_version" description:"集群版本"`
	KubeConfig     string `json:"kube_config" description:"KubeConfig文本"`
	Status         int    `json:"status" description:"集群状态"`
	*Base
}

// TableName 方法用于返回数据库表名
// 这是一个 K8sCluster 结构体的方法，用于定义该结构体对应的数据库表名
// 返回值为字符串类型，表示表名
func (k *K8sCluster) TableName() string { return "k8s_cluster" }

// Create 插入数据（依赖库侧唯一索引 uniq_cluster_name）
func (k *K8sCluster) Create(db *gorm.DB) error {
	tx := db.Create(k)
	if tx.Error != nil {
		global.Logger.Error("创建集群失败",
			zap.String("cluster_name", k.ClusterName),
			zap.Error(tx.Error),
		)
		return tx.Error
	}
	global.Logger.Info("创建集群成功",
		zap.Uint32("cluster_id", k.ID),
		zap.String("cluster_name", k.ClusterName),
	)
	return nil
}

// GetByName 根据集群名获取集群信息
func (k *K8sCluster) GetByName(db *gorm.DB) (*K8sCluster, error) {
	if k.ClusterName == "" {
		return nil, errorcode.ErrorClusterNotFound
	}
	var kc K8sCluster
	err := db.Where("cluster_name = ? AND is_del = 0", k.ClusterName).First(&kc).Error
	return &kc, err
}

// GetByID 通过ID获取集群信息
func (k *K8sCluster) GetByID(db *gorm.DB) (*K8sCluster, error) {
	var kc K8sCluster
	err := db.Where("id = ? AND is_del = 0", k.ID).First(&kc).Error
	return &kc, err
}

// List 列出集群信息（分页）
func (k *K8sCluster) List(db *gorm.DB, page, limit int) ([]*K8sCluster, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	startIndex := (page - 1) * limit

	var list []*K8sCluster
	q := db.Model(&K8sCluster{}).Where("is_del = 0")
	if k.ClusterName != "" {
		q = q.Where("cluster_name LIKE ?", "%"+k.ClusterName+"%")
	}
	err := q.Offset(startIndex).Limit(limit).Find(&list).Error
	return list, err
}

// Update 更新数据（values 可为 map[string]interface{} 或 struct）
// 会检查影响行数，返回 error 表示失败或未修改
func (k *K8sCluster) Update(db *gorm.DB, values interface{}) error {
	tx := db.Model(&K8sCluster{}).
		Where("id = ? AND is_del = 0", k.ID).
		Updates(values)

	// 先判断 SQL 执行错误
	if tx.Error != nil {
		return tx.Error
	}

	// 再判断是否有行受到影响
	if tx.RowsAffected == 0 {
		return errorcode.ErrorClusterUpdateFail
	}

	return nil
}

// Delete 软删除（将 is_del 置 1）
func (k *K8sCluster) Delete(db *gorm.DB) error {
	tx := db.Model(&K8sCluster{}).
		Where("id = ? AND is_del = 0", k.ID).
		Update("is_del", 1)

	if tx.Error != nil {
		// 记录日志
		global.Logger.Error("软删除集群失败",
			zap.Uint32("cluster_id", k.ID),
			zap.Error(tx.Error),
		)
		// 返回 error，交给上层处理
		return tx.Error
	}

	if tx.RowsAffected == 0 {
		msg := fmt.Errorf("集群不存在或已删除: id=%d", k.ID)
		global.Logger.Warn("软删除集群未生效",
			zap.Uint32("cluster_id", k.ID),
		)
		return msg
	}

	// 成功时也可以打印一条 Info 日志
	global.Logger.Info("软删除集群成功",
		zap.Uint32("cluster_id", k.ID),
	)

	return nil
}

// Save 保存数据（存在则更新，不存在则创建）
func (k *K8sCluster) Save(db *gorm.DB) error {
	return db.Save(k).Error
}
