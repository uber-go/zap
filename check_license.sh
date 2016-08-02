#!/bin/bash

text=`cat LICENSE.txt | head -1`

ERROR_COUNT=0
while read file
do
    head -1 ${file} | grep -q "${text}"
    if [ $? -ne 0 ]; then
        echo "$file is missing license header."
        ERROR_COUNT+=1
    fi
done < <(git grep -l "" | grep "\.go")

exit $ERROR_COUNT
