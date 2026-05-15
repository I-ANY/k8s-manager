package dao

import (
	"k8soperation/internal/app/models"
	"time"
)

// K8sClusterCreate 创建集群
func (d *Dao) K8sClusterCreate(cluster_name, cluster_version, kube_config string) error {
	nowTime := uint32(time.Now().Unix())
	kc := models.K8sCluster{
		ClusterName:    cluster_name,
		ClusterVersion: cluster_version,
		KubeConfig:     kube_config,
		Status:         0,
		Base: &models.Base{
			CreatedAt:  nowTime,
			ModifiedAt: nowTime,
			IsDel:      0,
		},
	}
	return kc.Create(d.db)
}

// K8sClusterGetByName 通过集群名称获取集群
func (d *Dao) K8sClusterGetByName(clusterName string) (*models.K8sCluster, error) {
	kc := models.K8sCluster{
		ClusterName: clusterName,
	}
	return kc.GetByName(d.db)
}

// K8sClusterGetByID 通过集群ID获取集群
func (d *Dao) K8sClusterGetByID(clusterId uint32) (*models.K8sCluster, error) {
	kc := models.K8sCluster{
		Base: &models.Base{
			ID: clusterId,
		},
	}
	return kc.GetByID(d.db)
}

// K8sClusterUpdate 更新集群
func (d *Dao) K8sClusterUpdate(id uint32, clusterName, clusterVersion, kubeConfig string, status int) error {
	nowTime := uint32(time.Now().Unix())
	kc := models.K8sCluster{
		Base: &models.Base{
			ID: id,
		},
	}
	values := map[string]interface{}{
		"cluster_name":    clusterName,
		"cluster_version": clusterVersion,
		"kube_config":     kubeConfig,
		"status":          status,
		"modified_at":     nowTime,
	}
	return kc.Update(d.db, values)
}

// K8sClusterList 列出集群信息
func (d *Dao) K8sClusterList(cluster_name string, page, limit int) ([]*models.K8sCluster, error) {
	kc := &models.K8sCluster{
		ClusterName: cluster_name,
	}
	return kc.List(d.db, page, limit)
}

// K8sClusterDelete 删除集群信息
func (d *Dao) K8sClusterDelete(clusterId uint32) error {
	kc := &models.K8sCluster{
		Base: &models.Base{
			ID: clusterId,
		},
	}
	return kc.Delete(d.db)
}
