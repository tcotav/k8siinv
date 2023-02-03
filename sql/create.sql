-- Create a user for the app
-- 
-- CREATE USER 'k8siinv'@'%' IDENTIFIED WITH mysql_native_password BY 'password';
--
-- CREATE DATABASE k8siinv;
-- 
-- GRANT ALL PRIVILEGES ON k8siinv.* TO 'imginv'@'%';

USE k8siinv;

-- cluster info
CREATE TABLE clusters (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name varchar(256) NOT NULL,
    firstSeen TIMESTAMP NOT NULL,
    lastSeen TIMESTAMP NOT NULL, 
    recordDate TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=INNODB;
CREATE UNIQUE INDEX ix_clusters_name
ON clusters (name);

-- image info itself
CREATE TABLE images (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name varchar(256) NOT NULL,
    version varchar(256) NOT NULL,
    firstSeen TIMESTAMP NOT NULL,
    lastSeen TIMESTAMP NOT NULL, 
    recordDate TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=INNODB;
CREATE UNIQUE INDEX ix_images_name_ver
ON images (name, version);

-- metadata related to how the image was used and when 
CREATE TABLE clusterinventory (
    id INT AUTO_INCREMENT PRIMARY KEY,
    imgid INT NOT NULL,
    clusterid INT NOT NULL,
    podname varchar(256) NOT NULL,
    namespace varchar(256) NOT NULL,
    startTime timestamp NOT NULL,
    recordDate TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (imgid)
        REFERENCES images (id),
    FOREIGN KEY (clusterid)
        REFERENCES clusters (id)
) ENGINE=INNODB;

-- activity
CREATE TABLE activity (
    id INT AUTO_INCREMENT PRIMARY KEY,
    clusterid int NOT NULL,
    activityDate timestamp NOT NULL,
    recordDate TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    version varchar(10) NOT NULL,
    FOREIGN KEY (clusterid)
        REFERENCES clusters (id)
) ENGINE=INNODB;
CREATE UNIQUE INDEX ix_activity_clusterid_actdate
ON activity (clusterid, activityDate);