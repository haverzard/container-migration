#!/bin/bash
url=$1
upper_bound="${2:=1}"
lower_bound="${3:=1}"
init_replicas="${4:=1}"
total_jobs="${5:=3}"

if [[$upper_bound < $lower_bound]] then
  exit
fi

for id in {1..$total_jobs}
do
cat <<EOT >> experiments/jobs/job$id.yaml
apiVersion: kubeflow.org/v1
kind: TFJob
metadata:
  name: job$id
spec:
  max-instances: $upper_bound
  min-instances: $lower_bound
  cleanPodPolicy: "All"
  tfReplicaSpecs:
    PS:
      replicas: 1
      restartPolicy: OnFailure
      template:
        spec:
          terminationGracePeriodSeconds: 0
          containers:
          - name: tensorflow
            image: tensorflow/tensorflow:1.15.0-py3
            command: ["/bin/bash", "-c", "curl -s $url/mnist-df.py | python3 -"]
            ports:
            - containerPort: 2222
              name: tfjob-port
            resources:
              requests:
                cpu: "500m"
                memory: "512Mi"
              limits:
                cpu: "1"
                memory: "1Gi"
    Worker:
      replicas: $init_replicas
      restartPolicy: OnFailure
      template:
        spec:
          terminationGracePeriodSeconds: 0
          containers:
          - name: tensorflow
            image: tensorflow/tensorflow:1.15.0-py3
            command: ["/bin/bash", "-c", "curl -s $url/mnist-df.py | python3 -"]
            env:
            - name: "global_steps"
              value: "10000"
            ports:
            - containerPort: 2222
              name: tfjob-port
            resources:
              requests:
                cpu: "500m"
                memory: "512Mi"
              limits:
                cpu: "1"
                memory: "1Gi"
EOT
done

