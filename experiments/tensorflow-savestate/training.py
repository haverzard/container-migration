import argparse
import sys
import os
import ast

from tensorflow.examples.tutorials.mnist import input_data

mnist = input_data.read_data_sets("MNIST_data/", one_hot=True)

import tensorflow as tf


class Empty:
    pass


FLAGS = Empty()


def main(_):
    ps_hosts = FLAGS.ps_hosts.split(",")
    worker_hosts = FLAGS.worker_hosts.split(",")

    # Create a cluster from the parameter server and worker hosts.
    cluster = tf.train.ClusterSpec({"ps": ps_hosts, "worker": worker_hosts})

    # Create and start a server for the local task.
    server = tf.distribute.Server(
        cluster, job_name=FLAGS.job_name, task_index=FLAGS.task_index
    )
    if FLAGS.job_name == "ps":
        server.join()
    elif FLAGS.job_name == "worker":
        print("WORKER RUN")
        # Assigns ops to the local worker by default.
        with tf.device(
            tf.train.replica_device_setter(
                worker_device="/job:worker/task:%d" % FLAGS.task_index, cluster=cluster
            )
        ):

            # Build model...
            x = tf.placeholder(tf.float32, [None, 784])
            W = tf.Variable(tf.zeros([784, 10]))
            b = tf.Variable(tf.zeros([10]))
            y = tf.nn.softmax(tf.matmul(x, W) + b)
            y_ = tf.placeholder(tf.float32, [None, 10])
            cross_entropy = tf.reduce_mean(
                -tf.reduce_sum(y_ * tf.log(y), reduction_indices=[1])
            )
            learning_rate = 0.05
            global_step = tf.train.get_or_create_global_step()
            train_step = tf.train.GradientDescentOptimizer(learning_rate).minimize(
                cross_entropy, global_step=global_step
            )

        # The StopAtStepHook handles stopping after running given steps.
        saver = tf.train.Saver(max_to_keep=10)
        scaffold = tf.train.Scaffold(saver=saver)
        hooks = [
            tf.train.StopAtStepHook(last_step=FLAGS.global_steps),
            # tf.estimator.CheckpointSaverHook(
            #     FLAGS.train_dir, save_steps=10, scaffold=scaffold
            # ),
        ]

        # The MonitoredTrainingSession takes care of session initialization,
        # restoring from a checkpoint, saving to a checkpoint, and closing when done
        # or an error occurs.
        with tf.train.MonitoredTrainingSession(
            master=server.target,
            is_chief=(FLAGS.task_index == 0),
            config=tf.ConfigProto(
                device_filters=["/job:ps", "/job:worker/task:%d" % FLAGS.task_index]
            ),
            save_checkpoint_steps=10,
            scaffold=scaffold,
            checkpoint_dir=FLAGS.train_dir,
            hooks=hooks,
        ) as mon_sess:
            ckpt = tf.train.get_checkpoint_state(FLAGS.train_dir)
            # if ckpt and ckpt.model_checkpoint_path:
            #     # Restores from checkpoint
            #     saver.restore(mon_sess, ckpt.model_checkpoint_path)
            while not mon_sess.should_stop():
                batch_xs, batch_ys = mnist.train.next_batch(16)
                _, step = mon_sess.run(
                    [train_step, global_step], feed_dict={x: batch_xs, y_: batch_ys}
                )
            print(os.listdir(FLAGS.train_dir))
            # exit()
            # sys.stderr.write('global_step: '+str(step))
            # sys.stderr.write('\n')


if __name__ == "__main__":
    TF_CONFIG = ast.literal_eval(os.environ["TF_CONFIG"])
    FLAGS.job_name = TF_CONFIG["task"]["type"]
    FLAGS.task_index = TF_CONFIG["task"]["index"]
    FLAGS.ps_hosts = ",".join(TF_CONFIG["cluster"]["ps"])
    FLAGS.train_dir = "/cp/"
    FLAGS.worker_hosts = ",".join(TF_CONFIG["cluster"]["worker"])
    FLAGS.global_steps = (
        int(os.environ["global_steps"]) if "global_steps" in os.environ else 100000
    )
    tf.compat.v1.app.run(main=main, argv=[sys.argv[0]])
