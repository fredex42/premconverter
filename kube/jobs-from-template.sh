#!/usr/bin/env bash

NUM_SPLITS=5
REAL_INPUT_PATH=/path/to/projects
REAL_OUTPUT_PATH=/path/to/upgraded
LISTFILE=dev-environment-projects

for i in `seq 1 $NUM_SPLITS`; do
    cat templates/RunAsJob.yaml | sed s/\{\{index\}\}/$i/g > /tmp/1
    cat /tmp/1 | sed s/\{\{listfile\}\}/$LISTFILE/g > /tmp/2
    cat /tmp/2 | sed 's.{{real-input-path}}.'$REAL_INPUT_PATH.g > /tmp/3
    cat /tmp/3 | sed 's.{{real-output-path}}.'$REAL_OUTPUT_PATH.g > premconverter-job-$i.yaml
done

rm -f /tmp/1 /tmp/2 /tmp/3
