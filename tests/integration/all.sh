#!/bin/sh

reset
rm -rf images/* 

while LFS= read n; do 
	./run.sh $n </dev/null 2>&1 | tee $(echo "$n" | sed 's/ /_/g').log || true
done < configurations-list.txt
