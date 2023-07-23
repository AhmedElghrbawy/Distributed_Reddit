#!/bin/bash

nshards=$(jq -r '.number_of_shards' ../config.json)
nreplicas=$(jq -r '.number_of_replicas' ../config.json)

session="rdb"

tmux start-server

tmux new-session -d -s $session

for (( i=0; i<$nshards; i++ ))
do 
    tmux new-window -n S$i

    for (( j=1; j<$nreplicas; j++ ))
    do
        tmux split-window -h 
    done

    tmux select-layout even-horizontal

    for (( j=0; j<$nreplicas; j++ ))
    do
        tmux select-pane -t $j
        tmux send-keys "go run cmd/replicated_db/*.go $i $j" C-m 
    done

done



# kill the default window created from new-session command
tmux kill-window -t $session:0

tmux select-window -t $session:S0

tmux attach-session -t $session
