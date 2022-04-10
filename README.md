# container-migration

## Overview

Fork from [NTHU-LSALAB/DRAGON](https://github.com/NTHU-LSALAB/DRAGON).

Modification of dynamic resource scheduler for deep learning to support container migration policy.


It consists of:

Component         | Description
------------------|---
[DRAGON](https://github.com/haverzard/container-migration/tree/main/cmd/DRAGON) | DRAGON is a dynamic resource scheduler for distributed parameter server Tensorflow training jobs (TFJob) with automatic scheduling and scaling strategies.
[Container Monitor](https://github.com/haverzard/container-migration/tree/main/internal/container-monitor) | Container Monitor is a monitoring module for training tasks. It's deployed using DaemonSet to localize the monitoring in order to reduce the overhead.
[Preemptible Job](https://github.com/haverzard/container-migration/tree/main/internal/jobs) | A modified distributed Tensorflow job to support migration.

## Project Structure

```
.
├── Makefile                  # Makefile command helpers
├── go.mod                    # DRAGON dependencies
├── cmd                       
│   └── DRAGON                # DRAGON main application
├── deployments
│   ├── docker                # Dockerfiles for DRAGON, Container Monitor, and custom TF image
│   ├── kubernetes            # Kubernetes config files for `dragon-tf-operator`, `container-monitor`, and TFJob
│   └── terraform             # Terraform definitions with GCP provider
├── internal
│   ├── jobs                  # Python Distributed Parameter Server Tensorflow Job
│   └── container-monitor     # Container Monitor project folder
├── pkg                       # DRAGON package
│   ├── apis
│   ├── backend
│   ├── client
│   ├── common
│   ├── control
│   ├── controller.v1         # scheduling policies
│   ├── logger
│   ├── util
│   └── version
├── scripts
└── vendor 
```

## Requirements

### Cloud

- [GCP Account](https://console.cloud.google.com/)
- [Terraform v.1.x](terraform.io)

### On-Premise

- [On-Premise Kubernetes v.1.19.x](https://kubernetes.io/docs/tasks/tools/)

## Installation

### Setup Cloud Cluster (GCP only)

The GKE cluster has been defined on Terraform. To initialize it, run the command below:

```sh
cd deployments/terraform
terraform apply
```

### Build Custom Docker Images

1.  Make you have logged in to your docker registry.

    ```sh
    docker login
    ```

2.  Change the make commands for `release-dragon`, `release-api`, and `release-tf-image` in `Makefile` so it points to your Docker repositories.

3. Do not forget to update images on the Kubernetes config files in `deployments/kubernetes/` folder.

### Deploy DRAGON and Container Monitor

***
Note DRAGON and Container Monitor have been built and uploaded in the Docker Hub. If you want to change/modify it, please refer to `Build Custom Docker Images` step.
***

Run the command below to install the modified DRAGON with container migration:
```sh
make install-custom
```

Run the command below to install the original DRAGON:
```sh
make install
```

## Scheduling and Scaling Strategies

### DRAGON Default Strategies

* DRAGON tries to schedule pending jobs in the FIFO order and ignores the job which cannot meet its resource requirements.
* If there exist some jobs which have been waiting for more than 30 seconds, pick the longest one. DRAGON will schedule the job if it can meet its resource requirements after scaling down running jobs. This is for higher system throughput.
* If there exists some idle resources and DRAGON didn't perform any scheduling and scaling actions for more than one minute, DRAGON tries to scale up running jobs. This is for higher resource utilization.

* DRAGON prefers to schedule all replicas within one node due to communication overhead.
* When performing scaling down, DRAGON prefers to terminate the lonely replicas first. (lonely means that the location of the replica is different from the parameter server)
* When performing scaling up, DRAGON prefers to schedule the new replicas to where the parameter server is located.

### Migration Strategy
* Container Monitor will listen for evaluation results from the training tasks and asynchronously decides if a migration must happen based on the result and the resource usage distribution.
* When the resource usage distribution is imbalanced or a task has converged, Container Monitor will decide to migrate the task.
* When a task is decided to be migrated, Container Monitor will a migration request to DRAGON through the migration microservice.
* DRAGON will receive the migration request and enqueue it as a migration job.
* DRAGON will migrate the tasks using the migration policy by destroying the Pod and recreating it in another Node.
* DRAGON will ignore the migration job if the target Node is the same as the source Node.

## Questions

Please contact me on [email](mailto:yonatanviody@gmail.com) if you have any questions 😊
