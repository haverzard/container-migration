import { ContainerCategory } from "../models/category";
import { Metric } from '../models/evaluation';
import Queue from './queue';

const TIME_WEIGHT = parseFloat(process.env.TIME_WEIGHT || "100.0")
const METRIC_WEIGHT = parseFloat(process.env.METRIC_WEIGHT || "20.0")
const PROGRESSING_THREESHOLD = parseFloat(process.env.PROGRESSING_THREESHOLD || "10")
const CONVERGED_THREESHOLD = parseFloat(process.env.CONVERGED_THREESHOLD || "-25")


interface ModelEvaluation {
  metric: Metric
  time: number
}

export class Pod {
  name: string
  score: number
  accessedAt: number
  category: ContainerCategory
  evaluations: Queue<ModelEvaluation>

  constructor(name: string) {
    this.name = name
    this.score = 0
    this.accessedAt = Date.now()
    this.category = ContainerCategory.Progressing
    this.evaluations = new Queue<ModelEvaluation>()
  }

  speculate() {
    let size = this.evaluations.size
    if (size < 2) {
      return;
    }

    let oldEvaluation = this.evaluations.seek()!
    while (--size) {
      oldEvaluation = this.evaluations.pop()!
    }

    let newEvaluation = this.evaluations.seek()!
    let dt = newEvaluation.time - oldEvaluation.time
    let dmetric = newEvaluation.metric.value - oldEvaluation.metric.value
    let score = (dmetric * METRIC_WEIGHT - dt / TIME_WEIGHT)

    let oldCategory = this.category
    let dscore = score - this.score
    console.log("time", dt)
    console.log("SCORE", score, dscore)
    if (dscore > PROGRESSING_THREESHOLD || dscore > 5) {
      this.category = ContainerCategory.Progressing
    } else if (dscore < -5) {
      if (this.category == ContainerCategory.Progressing && dscore >= CONVERGED_THREESHOLD) {
        this.category = ContainerCategory.Watching
      } else {
        this.category = ContainerCategory.Converged
      }
    }
    if (oldCategory != this.category) {
      this.score = score
    }
  }

  addEvaluation(evaluation: ModelEvaluation) {
    this.evaluations.push(evaluation)
  }
}

export class PodStorage {
  private pods: { [key: string]: Pod }

  constructor() {
    this.pods = {}
  }

  findPod(name: string) {
    let pod = this.pods[name]
    if (pod) {
      pod.accessedAt = Date.now()
    }
    return pod
  }

  addPod(name: string) {
    const pod = new Pod(name)
    this.pods[name] = pod
    return pod
  }

  garbageCollection() {
    let time = Date.now()
    let pods = { ...this.pods }
    for (let pod in pods) {
      if ((time - pods[pod].accessedAt) >= 3600000) { // 1 hour period
        delete this.pods[pod]
      }
    }
  }
}

const podStorage = new PodStorage()

export default podStorage;