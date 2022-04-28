#!/bin/bash
# author:ipso
# blog:www.ipso.live

modified=git status
newfile=$modified | grep "new file"
modified=$modified | grep "modified"
if [ $newfile!="" ]
then
    git add .
    git commit -m "$1"
    git push origin master
elif [ modified!="" ]
then
    git commit -m "$1"
    git push origin master
else
    git status >> logs.text
fi