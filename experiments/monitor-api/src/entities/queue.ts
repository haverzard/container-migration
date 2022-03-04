class LinkedNode<T> {
  value: T;
  next: LinkedNode<T> | null;
  prev: LinkedNode<T> | null;

  constructor(value: T) {
    this.value = value;
    this.next = null;
    this.prev = null;
  }
}

export default class Queue<T> {
  head: LinkedNode<T> | null
  tail: LinkedNode<T> | null
  size: number

  constructor() {
    this.head = null
    this.tail = null
    this.size = 0
  }

  push(value: T) {
    let node = new LinkedNode(value);
    if (!this.tail) {
      this.tail = node;
      this.head = node;
    } else {
      this.tail.next = node;
      node.prev = this.tail;
      this.tail = node;
    }
    this.size++;
  }

  pop() {
    if (!this.head) {
      return null
    }

    let node = this.head;
    if (node == this.tail) {
      this.tail = null;
    }

    let next = this.head.next;
    if (next) {
      next.prev = null;
    }
    this.head = next;
    this.size--;
    return node.value;
  }

  seek() {
    if (!this.head) {
      return null
    }

    return this.head.value;
  }
}