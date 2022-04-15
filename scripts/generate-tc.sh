#!/bin/bash
url="${1:-https://lsalab.cs.nthu.edu.tw/~ericyeh/DRAGON}"
upper_bound="${2:-1}"
lower_bound="${3:-1}"
init_replicas="${4:-1}"
output_file="${5:-deployments/kubernetes/jobs/example.yaml}"

cat <<EOT > $output_file
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
            image: haverzard/tf-image:$TF_IMAGE_VERSION
            command: ["/bin/bash", "-c", "curl -s $url/mnist-df.py > tf.py && (python tf.py chief & python tf.py)"]
            env:
            - name: "global_steps"
              value: "500"
            - name: "batch_interval"
              value: "20"
            - name: "max_workers"
              value: "$upper_bound"
            ports:
            - containerPort: 2222
              name: tfjob-port
            resources:
              requests:
                cpu: "500m"
                memory: "1Gi"
              limits:
                cpu: "1"
                memory: "2Gi"
    Worker:
      replicas: $init_replicas
      restartPolicy: OnFailure
      template:
        spec:
          terminationGracePeriodSeconds: 0
          containers:
          - name: tensorflow
            image: haverzard/tf-image:$TF_IMAGE_VERSION
            command: ["/bin/bash", "-c", "curl -s $url/mnist-df.py | python3 -"]
            env:
            - name: "global_steps"
              value: "500"
            - name: "batch_interval"
              value: "20"
            - name: "max_workers"
              value: "$upper_bound"
            - name: NODE_IP
              valueFrom:
                fieldRef:
                  fieldPath: status.hostIP
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            ports:
            - containerPort: 2222
              name: tfjob-port
            resources:
              requests:
                cpu: "500m"
                memory: "1Gi"
              limits:
                cpu: "1"
                memory: "2Gi"
EOT
