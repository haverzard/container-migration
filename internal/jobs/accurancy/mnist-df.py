from tensorflow.examples.tutorials.mnist import input_data
import tensorflow as tf
import sys
import os
import ast


class Empty:
    pass


def batch_norm(x, axes=[0, 1, 2]):
    """
    Batch Normalization layer with `tf.nn`

    Args:
        x       : input neurons
        axes    : axes to compute mean and std. Use [0,1,2] for 2D norm and [0] for 1D norm
    """
    mean_x, std_x = tf.nn.moments(x, axes=axes, keep_dims=True)
    return tf.nn.batch_normalization(x, mean_x, std_x, 0, 1, 1e-3)


def conv2d(
    x, filters=None, kernel_size=None, n_inputs=None, strides=(1, 1), name="default"
):
    """
    Convolution 2D layer with `tf.nn`

    Args:
        x           : input neurons
        filters     : number of filters
        kernel_size : 2D kernel size
        n_inputs    : number of inputs
        strides     : 2D stride size
        name        : layer name
    """
    W = tf.get_variable(
        "W" + name,
        shape=(*kernel_size, n_inputs, filters),
        initializer=tf.contrib.layers.xavier_initializer(),
    )
    b = tf.get_variable(
        "B" + name, shape=(filters), initializer=tf.contrib.layers.xavier_initializer()
    )
    # Conv2D wrapper, with bias and relu activation
    x = tf.nn.conv2d(x, W, strides=[1, *strides, 1], padding="VALID")
    x = tf.nn.bias_add(x, b)
    x = batch_norm(x)
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
        mnist = input_data.read_data_sets("MNIST_data/", one_hot=True)
        mnist.train._images = mnist.train._images.reshape(-1, 28, 28, 1)
        mnist.test._images = mnist.test._images.reshape(-1, 28, 28, 1)

        # Assigns ops to the local worker by default.
        with tf.device(
            tf.train.replica_device_setter(
                worker_device="/job:worker/task:%d" % FLAGS.task_index, cluster=cluster
            )
        ):

            # Build model...
            x = tf.placeholder(tf.float32, [None, 28, 28, 1])
            y_exp = tf.placeholder(tf.float32, [None, 10])

            # Convolution layers
            conv1 = conv2d(x, n_inputs=1, filters=32, kernel_size=(5, 5), name="conv1")
            conv2 = conv2d(
                conv1, n_inputs=32, filters=64, kernel_size=(5, 5), name="conv2"
            )
            conv3 = conv2d(
                conv2, n_inputs=64, filters=96, kernel_size=(5, 5), name="conv3"
            )
            conv4 = conv2d(
                conv3, n_inputs=96, filters=128, kernel_size=(5, 5), name="conv4"
            )
            conv5 = conv2d(
                conv4, n_inputs=128, filters=160, kernel_size=(5, 5), name="conv5"
            )

            # Fully connected layers
            fc1 = tf.reshape(conv5, [-1, 10240])

            Wfc1 = tf.get_variable(
                "Wfc1",
                shape=(10240, 10),
                initializer=tf.contrib.layers.xavier_initializer(),
            )
            bfc1 = tf.get_variable(
                "Bfc1", shape=(10), initializer=tf.contrib.layers.xavier_initializer()
            )
            fc2 = tf.add(tf.matmul(fc1, Wfc1), bfc1)
            fc2 = batch_norm(fc2, axes=[0])

            y_obv = tf.nn.softmax(fc2)

            # Evaluation and other operations
            cross_entropy = tf.reduce_mean(
                -tf.reduce_sum(y_exp * tf.log(y_obv), reduction_indices=[1])
            )
            learning_rate = 0.001
            global_step = tf.train.get_or_create_global_step()
            train_step = tf.train.GradientDescentOptimizer(learning_rate).minimize(
                cross_entropy, global_step=global_step
            )
            _, acc_op = tf.metrics.accuracy(
                labels=tf.argmax(y_exp, axis=1), predictions=tf.argmax(y_obv, 1)
            )

        # The StopAtStepHook handles stopping after running given steps.
        # The FinalOpsHook handles running the final evaluation after the training is stopped
        is_chief = FLAGS.task_index == 0
        final_ops = tf.train.FinalOpsHook(
            acc_op, {x: mnist.test._images, y_exp: mnist.test._labels}
        )
        hooks = [tf.train.StopAtStepHook(last_step=FLAGS.global_steps)]
        if is_chief:
            hooks.append(final_ops)

        # The MonitoredTrainingSession takes care of session initialization,
        # restoring from a checkpoint, saving to a checkpoint, and closing when done
        # or an error occurs.
        batches = 0
        with tf.train.MonitoredTrainingSession(
            master=server.target,
            is_chief=is_chief,
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

                # Evaluation Step
                batches += 1
                if not mon_sess.should_stop() and batches % FLAGS.batch_interval == 0:
                    batch_xs, batch_ys = mnist.test.next_batch(16)
                    accuracy = mon_sess.run(
                        acc_op, feed_dict={x: batch_xs, y_exp: batch_ys}
                    )
                    print(f"ACC {step} {accuracy}")
        print("DONE")
        if is_chief:
            print("Final Accuracy: " + str(final_ops._final_ops_values))


if __name__ == "__main__":
    TF_CONFIG = ast.literal_eval(os.environ["TF_CONFIG"])
    # Handle Chief/Coordinator process
    if len(sys.argv) == 2 and sys.argv[1] == "chief":
        FLAGS.job_name = "worker"
    else:
        FLAGS.job_name = TF_CONFIG["task"]["type"]
    # Task identity
    FLAGS.task_index = TF_CONFIG["task"]["index"]
    # Cluster config
    FLAGS.ps_hosts = ",".join(TF_CONFIG["cluster"]["ps"])
    FLAGS.worker_hosts = ",".join(TF_CONFIG["cluster"]["worker"])
    FLAGS.global_steps = (
        int(os.environ["global_steps"]) if "global_steps" in os.environ else 100000
    )
    # Evaluation config
    FLAGS.batch_interval = (
        int(os.environ["batch_interval"]) if "batch_interval" in os.environ else 20
    )
    tf.app.run(main=main, argv=[sys.argv[0]])
