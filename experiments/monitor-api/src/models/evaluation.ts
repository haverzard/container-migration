export interface Metric {
  name: string;
  value: number;
}

export interface Evaluation {
  name: string;
  metric: Metric;
}