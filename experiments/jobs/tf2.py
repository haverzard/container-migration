import argparse
import sys
import os
import ast
import time
import tensorflow as tf
import numpy as np
import requests

(x_train, y_train), (x_test, y_test) = tf.keras.datasets.mnist.load_data()
x_train = x_train.reshape(-1, x_train.shape[-1] * x_train.shape[-2])
y_train = np.eye(10)[y_train]

class Empty:
    pass


FLAGS = Empty()


tf.compat.v1.disable_eager_execution()

def dataset_fn(input_context):
  global_batch_size = 64
  batch_size = input_context.get_per_replica_batch_size(global_batch_size)

  x = tf.random.uniform((10, 10))
  y = tf.random.uniform((10,))

  dataset = tf.data.Dataset.from_tensor_slices((x, y)).shuffle(10).repeat()
  dataset = dataset.shard(
      input_context.num_input_pipelines,
      input_context.input_pipeline_id)
  dataset = dataset.batch(batch_size)
  dataset = dataset.prefetch(2)

  return dataset


def main(_):
    ps_hosts = FLAGS.ps_hosts.split(",")
    worker_hosts = FLAGS.worker_hosts.split(",")

    # Create a cluster from the parameter server and worker hosts.
    cluster = tf.compat.v1.train.ClusterSpec({"ps": ps_hosts, "worker": worker_hosts})

    # Create and start a server for the local task.
    server = tf.compat.v1.distribute.Server(
        cluster, job_name=FLAGS.job_name, task_index=FLAGS.task_index
    )
    cluster_resolver = tf.distribute.cluster_resolver.SimpleClusterResolver(
        cluster, rpc_layer="grpc", task_type=FLAGS.job_name, task_id=FLAGS.task_index
    )
    strategy = tf.distribute.experimental.ParameterServerStrategy(
        cluster_resolver
    )

    if FLAGS.job_name == "ps":
        server.join()
    elif FLAGS.job_name == "worker":

        # Assigns ops to the local worker by default.
        with tf.compat.v1.device(
            tf.compat.v1.train.replica_device_setter(
                worker_device="/job:worker/task:%d" % FLAGS.task_index,
                cluster=cluster,
            )
        ):

            # Build model...
            x = tf.compat.v1.placeholder(tf.float32, [None, 784])
            W = tf.compat.v1.Variable(tf.zeros([784, 10]))
            b = tf.compat.v1.Variable(tf.zeros([10]))
            y = tf.compat.v1.nn.softmax(tf.matmul(x, W) + b)
            y_ = tf.compat.v1.placeholder(tf.float32, [None, 10])
            cross_entropy = tf.compat.v1.reduce_mean(
                -tf.compat.v1.reduce_sum(
                    y_ * tf.compat.v1.log(y), reduction_indices=[1]
                )
            )
            learning_rate = 0.05
            global_step = tf.compat.v1.train.get_or_create_global_step()
            train_step = tf.compat.v1.train.GradientDescentOptimizer(
                learning_rate
            ).minimize(cross_entropy, global_step=global_step)
            acc, acc_op = tf.compat.v1.metrics.accuracy(
                labels=tf.argmax(y_, axis=1), predictions=tf.argmax(y, 1)
            )

        with strategy.scope():
            print(type(strategy), hasattr(strategy, "distribute_datasets_from_function"))
            dc = strategy.distribute_datasets_from_function(dataset_fn)
            print(dc)
        tf.compat.v1.enable_eager_execution()
        # The StopAtStepHook handles stopping after running given steps.
        hooks = [
            tf.compat.v1.train.StopAtStepHook(last_step=FLAGS.global_steps),
        ]

        # The MonitoredTrainingSession takes care of session initialization,
        # restoring from a checkpoint, saving to a checkpoint, and closing when done
        # or an error occurs.
        with tf.compat.v1.train.MonitoredTrainingSession(
            master=server.target,
            is_chief=(FLAGS.task_index == 0),
            config=tf.compat.v1.ConfigProto(
                device_filters=["/job:ps", "/job:worker/task:%d" % FLAGS.task_index]
            ),
            hooks=hooks,
        ) as mon_sess:
            while not mon_sess.should_stop():
                batch_xs, batch_ys = x_train[:16], y_train[:16]
                _, step = mon_sess.run(
                    [train_step, global_step], feed_dict={x: batch_xs, y_: batch_ys}
                )
                if not mon_sess.should_stop():
                    batch_xs, batch_ys = x_train[:16], y_train[:16]
                    accuracy = mon_sess.run(
                        acc_op, feed_dict={x: batch_xs, y_: batch_ys}
                    )
                    sys.stderr.write("accuracy: " + str(accuracy) + "\n")
                    # requests.post(
                    #     FLAGS.monitoring_api + "/monitor",
                    #     headers={"Content-Type": "application/json"},
                    #     json={"pod": FLAGS.pod_name, "value": float(accuracy)},
                    # )
            # sys.stderr.write('global_step: '+str(step))
            # sys.stderr.write('\n')


if __name__ == "__main__":
    TF_CONFIG = ast.literal_eval(os.environ["TF_CONFIG"])
    FLAGS.job_name = TF_CONFIG["task"]["type"]
    FLAGS.task_index = TF_CONFIG["task"]["index"]
    FLAGS.ps_hosts = ",".join(TF_CONFIG["cluster"]["ps"])
    FLAGS.checkpoint_dir = "/cp/"
    FLAGS.checkpoint_basename = "model.ckpt"
    FLAGS.worker_hosts = ",".join(TF_CONFIG["cluster"]["worker"])
    FLAGS.global_steps = (
        int(os.environ["global_steps"]) if "global_steps" in os.environ else 100000
    )
    FLAGS.monitoring_api = (
        (f"http://{os.environ['NODE_IP']}:8081") if "NODE_IP" in os.environ else None
    )
    FLAGS.pod_name = os.getenv("POD_NAME", None)
    tf.compat.v1.app.run(main=main, argv=[sys.argv[0]])
