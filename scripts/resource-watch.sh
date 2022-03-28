#!/bin/bash
x=1
while true
do
    printf "\nWatching resource & pods on interval $x\n"
    printf "\nResource Usage\n"
    kubectl top nodes

    printf "\nPods\n"
    kubectl get pods -o wide

    sleep 5
    x=$(( $x + 1 ))
done