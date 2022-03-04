import { Router, Request, Response } from 'express';
import { Evaluation } from '../models/evaluation';
import podStorage, { Pod } from '../entities/pod';
import { ContainerCategory } from '../models/category';
import axios from 'axios';

const counter = {
  [ContainerCategory.Converged]: 0,
  [ContainerCategory.Progressing]: 0,
  [ContainerCategory.Watching]: 0,
}

const SERVER_ENDPOINT = process.env.SERVER_ENDPOINT || ""
const NODE_NAME = process.env.NODE_NAME || ""

async function decideMigration(pod: Pod) {
  // counter[pod.category]--;
  pod.speculate();
  // counter[pod.category]++;
  console.log("category", pod.category);
  const total = counter[ContainerCategory.Progressing] + counter[ContainerCategory.Watching];
  if (pod.category == ContainerCategory.Converged
    && total >= 0) {
    // Communicate with server endpoint
    console.log("hiii")
    await axios.post(SERVER_ENDPOINT + "/hello", { "pod": pod.name, "node": NODE_NAME })
  }
  podStorage.garbageCollection();
}

function monitorEvaluation(req: Request, res: Response) {
  const { name, metric } = req.body as Evaluation;
  let pod = podStorage.findPod(name);
  if (!pod) {
    pod = podStorage.addPod(name);
  }
  pod.addEvaluation({ metric: metric, time: Date.now() });
  res.send("Ok");

  decideMigration(pod);
}

const router = Router();

/**
 * @openapi
 * /api/monitor/:
 *   post:
 *     description: Store your model's evaluation
 *     consumes:
 *       - application/json
 *     requestBody:
 *       content:
 *         application/json:
 *           name: evaluation
 *           description: 
 *           schema:
 *             type: object
 *             required:
 *               - name
 *               - metric
 *             properties:
 *               name:
 *                 type: string
 *               metric:
 *                 type: object
 *                 properties:
 *                   name:
 *                     type: string
 *                   value:
 *                     type: integer
 *     responses:
 *       201:
 *         description: Ok
 */
router.post("/", monitorEvaluation);

export default router;
