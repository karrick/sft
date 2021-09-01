# BENCH_FORMAT=RFC1123Z
BENCH_FORMAT=RFC3339Nano

HYPERFINE_FORMAT='%a %A %b %B %c %C %d %D %e %F %g %G %h %H %I %j %k %l %m %M %N %p %P %r %R %s %S %T %u %w %x %X %y %Y %z %Z %% %+ %1 %2 %3 %4 %n %t'
# HYPERFINE_FORMAT='%F %T'
# HYPERFINE_FORMAT='%a, %d %b %Y %T %z'

build: sft

bench: main_test.go append_test.go copy_test.go
	go test -bench=. -benchmem $^

append_test.go: sft
	./sft -extra -f appendTime -append -o $@ $(BENCH_FORMAT)

copy_test.go: sft
	./sft -extra -f copyTime -o $@ $(BENCH_FORMAT)

clean:
	rm -f append copy sft append.go copy.go append_test.go copy_test.go

hyperfine: append copy
	hyperfine './append' './copy'

test: gotest copytest

copytest: copy
	./$<

gotest: main_test.go append_test.go copy_test.go
	go test -v $^

sft: main.go
	go build -o $@ $^

append: append.go
	go build -o $@ $^

append.go: sft
	./sft -m -extra -f appendTime -append -o $@ $(HYPERFINE_FORMAT)

copy: copy.go
	go build -o $@ $^

copy.go: sft
	./sft -m -extra -f copyTime -o $@ $(HYPERFINE_FORMAT)

.PHONY: build bench clean copytest gotest hyperfine test
