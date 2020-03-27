#!/bin/bash
NEXT_WAIT_TIME=0
until say 'done' && [ $NEXT_WAIT_TIME -eq 4 ]; do
$(( NEXT_WAIT_TIME++ ))
sleep 5s 
done