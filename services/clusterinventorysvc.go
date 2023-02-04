package services

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/tcotav/k8siinv/types"
)

type ClusterInventoryService struct {
	db *sql.DB
}

func NewClusterInventoryService(db *sql.DB) *ClusterInventoryService {
	return &ClusterInventoryService{db: db}
}

func (c *ClusterInventoryService) saveCluster(clusterName string, generatedAt string) (int64, error) {
	// couple parts -- want to save the cluster image metadata
	stmt, err := c.db.Prepare("INSERT INTO clusters(name, firstSeen, lastSeen) VALUES(?,?,?) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id), lastSeen=?")
	if err != nil {
		return -1, fmt.Errorf("prepare insert cluster fail %v", err.Error())
	}
	defer stmt.Close()

	res, err := stmt.Exec(clusterName, generatedAt, generatedAt, generatedAt)
	if err != nil {
		return -1, fmt.Errorf("exec insert cluster fail %v", err.Error())
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("lastinsertid - insert cluster fail %v", err.Error())
	}
	return lastId, nil
}

func (c *ClusterInventoryService) saveImage(name string, version string, generatedAt string) (int64, error) {
	// couple parts -- want to save the cluster image metadata
	stmt, err := c.db.Prepare("INSERT INTO images(name, version, firstSeen, lastSeen) VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id), lastSeen=?")
	if err != nil {
		return -1, fmt.Errorf("prepare - insert image fail %v", err.Error())
	}
	defer stmt.Close()
	res, err := stmt.Exec(name, version, generatedAt, generatedAt, generatedAt)
	if err != nil {
		return -1, fmt.Errorf("exec - insert image fail %v", err.Error())
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("lastinsertid - insert image fail %v", err.Error())
	}
	return lastId, nil
}

func (c *ClusterInventoryService) saveImageList(clusterId int64, generatedAt string, imgList []types.PodImageState) error {
	for _, pimg := range imgList {
		// for every image in the list, we do this
		for _, i := range pimg.Images {
			// check our image table and update as applicable
			imgParts := strings.Split(i, ":")
			version := ""
			if len(imgParts) == 2 {
				version = imgParts[1]
			}
			imgId, err := c.saveImage(imgParts[0], version, generatedAt)
			// update cluster image which details more of how and where our image is used
			stmt, err := c.db.Prepare("INSERT INTO clusterinventory(imgid, clusterid, podname, namespace, starttime) VALUES(?,?,?,?,?)")
			if err != nil {
				return fmt.Errorf("prepare clusterinv fail %v", err.Error())
			}
			defer stmt.Close()
			_, err = stmt.Exec(imgId, clusterId, pimg.Name, pimg.Namespace, pimg.StartTime)
			if err != nil {
				return fmt.Errorf("exec - prepare clusterinv fail %v", err.Error())
			}
		}
	}
	return nil
}

func (c *ClusterInventoryService) SaveClusterInventory(inv *types.ClusterInventory) error {
	// first get the clusterID or insert the cluster
	clusterId, err := c.saveCluster(inv.ClusterName, inv.GeneratedAt)
	if err != nil {
		return err
	}

	// then update the activity table showing this job just ran
	stmt, err := c.db.Prepare("INSERT INTO activity(clusterid, version, activityDate) VALUES(?, ?, ?)")
	if err != nil {
		return fmt.Errorf("prepare - insert activity fail %v", err.Error())
	}
	defer stmt.Close()
	_, err = stmt.Exec(clusterId, inv.Version, inv.GeneratedAt)
	if err != nil {
		return fmt.Errorf("exec - insert activity fail %v", err.Error())
	}
	// then update the clusterinv + images tables with our list of images
	err = c.saveImageList(clusterId, inv.GeneratedAt, inv.ImageState)
	return err
}
