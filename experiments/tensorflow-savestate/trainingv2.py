import argparse
import sys
import os
import ast
import tensorflow as tf
from tensorflow.core.util.event_pb2 import SessionLog
from tensorflow.python.training.summary_io import SummaryWriterCache
import numpy as np

(x_train, y_train), (x_test, y_test) = tf.keras.datasets.mnist.load_data()
x_train = x_train.reshape(-1, x_train.shape[-1] * x_train.shape[-2])
y_train = np.eye(10)[y_train]


class Empty:
    pass


FLAGS = Empty()

tf.compat.v1.disable_eager_execution()


def save(session, step, saver, summary_writer):
    """Saves the latest checkpoint, returns should_stop."""
    # logging.info("Calling checkpoint listeners before saving checkpoint %d...", step)
    # for l in self._listeners:
    #     l.before_save(session, step)

    # logging.info("Saving checkpoints for %d into %s.", step, self._save_path)
    save_path = os.path.join(FLAGS.checkpoint_dir, FLAGS.checkpoint_basename)
    saver.save(
        session,
        save_path,
        global_step=step,
        write_meta_graph=True,
    )
    summary_writer.add_session_log(
        SessionLog(status=SessionLog.CHECKPOINT, checkpoint_path=save_path),
        step,
    )
    # logging.info("Calling checkpoint listeners after saving checkpoint %d...", step)
    # should_stop = False
    # for l in self._listeners:
    #     if l.after_save(session, step):
    #         logging.info(
    #             "A CheckpointSaverListener requested that training be stopped. "
    #             "listener: {}".format(l)
    #         )
    #         should_stop = True
    # return should_stop


def main(_):
    ps_hosts = FLAGS.ps_hosts.split(",")
    worker_hosts = FLAGS.worker_hosts.split(",")

    # Create a cluster from the parameter server and worker hosts.
    cluster = tf.compat.v1.train.ClusterSpec({"ps": ps_hosts, "worker": worker_hosts})

    # Create and start a server for the local task.
    server = tf.compat.v1.distribute.Server(
        cluster, job_name=FLAGS.job_name, task_index=FLAGS.task_index
    )
    print("START")
    if FLAGS.job_name == "ps":
        server.join()
    elif FLAGS.job_name == "worker":
        print("WORKER RUN")
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

        # The StopAtStepHook handles stopping after running given steps.
        saver = tf.compat.v1.train.Saver(max_to_keep=10)
        summary_writer = SummaryWriterCache.get(FLAGS.checkpoint_dir)
        scaffold = tf.compat.v1.train.Scaffold(saver=saver)
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
            scaffold=scaffold,
            checkpoint_dir=FLAGS.checkpoint_dir,
            hooks=hooks,
        ) as mon_sess:
            while not mon_sess.should_stop():
                batch_xs, batch_ys = x_train[:16], y_train[:16]
                _, step = mon_sess.run(
                    [train_step, global_step], feed_dict={x: batch_xs, y_: batch_ys}
                )
                # save(mon_sess._sess._sess._sess._sess, step, saver, summary_writer)
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
    tf.compat.v1.app.run(main=main, argv=[sys.argv[0]])
