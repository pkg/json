BENCH_OPTS := \
	-test.run=xxx \
	-test.bench=CountWhitespace \
	-test.count=5 \
	-test.benchtime=5s

benchstat: old.txt new.txt
	benchstat {old,new}.txt

old.txt: json.old
	./$< $(BENCH_OPTS) | tee $@

new.txt: json.test
	./$< $(BENCH_OPTS) | tee $@
