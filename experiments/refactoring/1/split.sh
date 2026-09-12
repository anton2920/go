#!/bin/sh


grep 'before\.StatementOriginal' $1 | sed 's/before\.StatementOriginal/Statement/' >result-original.txt
grep 'before\.StatementBytesBuffer' $1 | sed 's/before\.StatementBytesBuffer/Statement/' >result-bytes-buffer.txt
grep 'before\.StatementMyFmt' $1 | sed 's/before\.StatementMyFmt/Statement/' >result-myfmt.txt
grep '1\.Statement' $1 | grep -v Statements | grep -v Prepared | sed 's/1\.//' >result-1.txt
grep '1\.StatementWithPreparedData' $1 | grep -v Statements | sed 's/StatementWithPreparedData/Statement/g' | sed 's/1\.//'  >result-1-prepared.txt
benchstat result-original.txt result-bytes-buffer.txt result-myfmt.txt result-1.txt result-1-prepared.txt >results.txt
rm -f result-original.txt result-bytes-buffer.txt result-myfmt.txt result-1.txt result-1-prepared.txt

grep 'before\.StatementsOriginal' $1 | sed 's/before\.StatementsOriginal/Statements/' >results-original.txt
grep 'before\.StatementsBytesBuffer' $1 | sed 's/before\.StatementsBytesBuffer/Statements/' >results-bytes-buffer.txt
grep 'before\.StatementsMyFmt' $1 | sed 's/before\.StatementsMyFmt/Statements/' >results-myfmt.txt
grep '1\.Statements' $1 | grep -v Prepared | sed 's/1\.//' >results-1.txt
grep '1\.StatementsWithPreparedData' $1 | sed 's/StatementsWithPreparedData/Statements/g' | sed 's/1\.//'  >results-1-prepared.txt
benchstat results-original.txt results-bytes-buffer.txt results-myfmt.txt results-1.txt results-1-prepared.txt >>results.txt
rm -f results-original.txt results-bytes-buffer.txt results-myfmt.txt results-1.txt results-1-prepared.txt
