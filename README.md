# k8siinv

Inventory DB of container images running on your k8s Clusters

## Basic Idea

Simple agent runs as a k8s cronjob on a cadence.  There it pulls all the container images currently being used on your cluster.  You run this on all clusters.  The client reports back to the central server where the information is stored in a database.

Client pulls: pod name, namespace, starttime, and list of images for all pods.  This is wrapped in a datastructure per cluster and timestamped.  You can see the [structs used here](types/clusterinventory.go).

Server side just dumps it to a database for later processing.  The [database schema is here](sql/create.sql).

### Why?

Maybe your org is perfect, but a lot of projects and their respective resources get abandoned.  You can track the age of a running container centrally and run an "old-timer report".

Also, if running container scans, there are tools for running them as part of your CI/CD, but fewer for running scans against images running in production.   You could pull your container list on cadence and scan each of the images against existing vulnerability scanning services.

Is anyone still using <container img :abc123>?  And why are our bills for container image registry storage so high?  Now you know.

Probably more.

### TODO

Lots.  This is mostly a spike.  Config files need to be added.  Currently it doesn't run inside of a cluster, but ideally this would run inside each cluster.  I'm testing each part manually running them with a kubeconfig handy.  We want this running internal so:

1) this should be put into helm, client and server
2) configmap created and integrated with the app
3) proper ClusterRoleBinding with adequate perms needs to be added with the service account user this will run as.
4) before helm, it needs to get shoved into containers
5) UI (I hate writing UIs so that'll be a while...)
6) sample useful report scripts



