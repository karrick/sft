# TEST_FORMAT='%a %A %b %B %c %C %d %D %e %F %g %G %h %H %I %j %k %l %m %M %N %p %P %r %R %s %S %T %u %w %x %X %y %Y %z %Z %% %+ %1 %2 %3 %4 %n %t'
# TEST_FORMAT='%F %T'
# TEST_FORMAT='%a, %d %b %Y %T %z'
TEST_FORMAT='RFC3339Nano'

build: sft

bench: main_test.go append_test.go copy_test.go
	go test -bench=. -benchmem $^

append_test.go: sft
	./sft -f appendTime -append -o $@ $(TEST_FORMAT)

copy_test.go: sft
	./sft -f copyTime -o $@ $(TEST_FORMAT)

clean:
	rm -f append copy sft append.go copy.go append_test.go copy_test.go

hyperfine: append copy
	hyperfine './append' './copy'

test: copy
	./$<

sft: main.go
	go build -o $@ $^

append: append.go
	go build -o $@ $^

append.go: sft
	./sft -m -f appendTime -append -o $@ $(TEST_FORMAT)

copy: copy.go
	go build -o $@ $^

copy.go: sft
	./sft -m -f copyTime -o $@ $(TEST_FORMAT)

.PHONY: build bench clean hyperfine test
