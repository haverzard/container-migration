import argparse
import sys
import os
import ast
import requests

from tensorflow.examples.tutorials.mnist import input_data

mnist = input_data.read_data_sets("MNIST_data/", one_hot=True)
mnist.train._images = mnist.train._images.reshape(-1, 28, 28, 1)
mnist.test._images = mnist.test._images.reshape(-1, 28, 28, 1)
import tensorflow as tf


class Empty:
    pass


def conv2d(x, W, b, strides=1):
    # Conv2D wrapper, with bias and relu activation
    x = tf.nn.conv2d(x, W, strides=[1, strides, strides, 1], padding="VALID")
    x = tf.nn.bias_add(x, b)
    mean_x, std_x = tf.nn.moments(x, axes=[0, 1, 2], keep_dims=True)
    x = tf.nn.batch_normalization(x, mean_x, std_x, 0, 1, 0.001)
    return tf.nn.relu(x)


FLAGS = Empty()


def main(_):
    ps_hosts = FLAGS.ps_hosts.split(",")
    worker_hosts = FLAGS.worker_hosts.split(",")

    # Create a cluster from the parameter server and worker hosts.
    cluster = tf.train.ClusterSpec({"ps": ps_hosts, "worker": worker_hosts})

    # Create and start a server for the local task.
    server = tf.train.Server(
        cluster, job_name=FLAGS.job_name, task_index=FLAGS.task_index
    )

    if FLAGS.job_name == "ps":
        server.join()
    elif FLAGS.job_name == "worker":

        # Assigns ops to the local worker by default.
        with tf.device(
            tf.train.replica_device_setter(
                worker_device="/job:worker/task:%d" % FLAGS.task_index, cluster=cluster
            )
        ):

            # Build model...
            x = tf.placeholder(tf.float32, [None, 28, 28, 1])
            y_exp = tf.placeholder(tf.float32, [None, 10])

            W0 = tf.get_variable(
                "W0",
                shape=(5, 5, 1, 32),
                initializer=tf.contrib.layers.xavier_initializer(),
            )
            W1 = tf.get_variable(
                "W1",
                shape=(5, 5, 32, 64),
                initializer=tf.contrib.layers.xavier_initializer(),
            )
            W2 = tf.get_variable(
                "W2",
                shape=(5, 5, 64, 96),
                initializer=tf.contrib.layers.xavier_initializer(),
            )
            W3 = tf.get_variable(
                "W3",
                shape=(5, 5, 96, 128),
                initializer=tf.contrib.layers.xavier_initializer(),
            )
            W4 = tf.get_variable(
                "W4",
                shape=(5, 5, 128, 160),
                initializer=tf.contrib.layers.xavier_initializer(),
            )
            W5 = tf.get_variable(
                "W5",
                shape=(10240, 10),
                initializer=tf.contrib.layers.xavier_initializer(),
            )
            b0 = tf.get_variable(
                "B0", shape=(32), initializer=tf.contrib.layers.xavier_initializer()
            )
            b1 = tf.get_variable(
                "B1", shape=(64), initializer=tf.contrib.layers.xavier_initializer()
            )
            b2 = tf.get_variable(
                "B2", shape=(96), initializer=tf.contrib.layers.xavier_initializer()
            )
            b3 = tf.get_variable(
                "B3", shape=(128), initializer=tf.contrib.layers.xavier_initializer()
            )
            b4 = tf.get_variable(
                "B4", shape=(160), initializer=tf.contrib.layers.xavier_initializer()
            )
            b5 = tf.get_variable(
                "B5", shape=(10), initializer=tf.contrib.layers.xavier_initializer()
            )

            conv1 = conv2d(x, W0, b0)
            conv2 = conv2d(conv1, W1, b1)
            conv3 = conv2d(conv2, W2, b2)
            conv4 = conv2d(conv3, W3, b3)
            conv5 = conv2d(conv4, W4, b4)

            fc1 = tf.reshape(conv5, [-1, 10240])
            fc1 = tf.add(tf.matmul(fc1, W5), b5)

            y_obv = tf.nn.softmax(fc1)
            cross_entropy = tf.reduce_mean(
                -tf.reduce_sum(y_exp * tf.log(y_obv), reduction_indices=[1])
            )
            learning_rate = 0.05
            global_step = tf.train.get_or_create_global_step()
            train_step = tf.train.GradientDescentOptimizer(learning_rate).minimize(
                cross_entropy, global_step=global_step
            )
            _, acc_op = tf.metrics.accuracy(
                labels=tf.argmax(y_exp, axis=1), predictions=tf.argmax(y_obv, 1)
            )

        # The StopAtStepHook handles stopping after running given steps.
        hooks = [tf.train.StopAtStepHook(last_step=FLAGS.global_steps)]

        # The MonitoredTrainingSession takes care of session initialization,
        # restoring from a checkpoint, saving to a checkpoint, and closing when done
        # or an error occurs.
        batches = 0
        with tf.train.MonitoredTrainingSession(
            master=server.target,
            is_chief=(FLAGS.task_index == 0),
            config=tf.ConfigProto(
                device_filters=["/job:ps", "/job:worker/task:%d" % FLAGS.task_index]
            ),
            hooks=hooks,
        ) as mon_sess:

            while not mon_sess.should_stop():
                batch_xs, batch_ys = mnist.train.next_batch(16)
                _, step = mon_sess.run(
                    [train_step, global_step], feed_dict={x: batch_xs, y_exp: batch_ys}
                )
                batches += 1
                if (
                    not mon_sess.should_stop()
                    and batches % FLAGS.batch_interval == 0
                    and (FLAGS.global_steps - step) < FLAGS.max_workers
                ):
                    batch_xs, batch_ys = mnist.test.next_batch(16)
                    accuracy = mon_sess.run(
                        acc_op, feed_dict={x: batch_xs, y_exp: batch_ys}
                    )
                    requests.post(
                        FLAGS.monitoring_api + "/monitor",
                        headers={"Content-Type": "application/json"},
                        json={"pod": FLAGS.pod_name, "value": float(accuracy)},
                    )
                # sys.stderr.write('global_step: '+str(step))
                # sys.stderr.write('\n')


if __name__ == "__main__":
    print("Test")
    TF_CONFIG = ast.literal_eval(os.environ["TF_CONFIG"])
    FLAGS.job_name = TF_CONFIG["task"]["type"]
    FLAGS.task_index = TF_CONFIG["task"]["index"]
    FLAGS.ps_hosts = ",".join(TF_CONFIG["cluster"]["ps"])
    FLAGS.worker_hosts = ",".join(TF_CONFIG["cluster"]["worker"])
    FLAGS.global_steps = (
        int(os.environ["global_steps"]) if "global_steps" in os.environ else 100000
    )
    FLAGS.monitoring_api = (
        (f"http://{os.environ['NODE_IP']}:8081") if "NODE_IP" in os.environ else None
    )
    FLAGS.pod_name = os.getenv("POD_NAME", None)
    FLAGS.max_workers = (
        int(os.getenv("max_workers")) if "max_workers" in os.environ else 10
    ) * 2
    FLAGS.batch_interval = (
        int(os.environ["batch_interval"]) if "batch_interval" in os.environ else 5
    )
    tf.app.run(main=main, argv=[sys.argv[0]])
