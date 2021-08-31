FORMAT_A='%a %A %b %B %c %C %d %D %e %F %g %G %h %H %I %j %k %l %m %M %N %p %P %r %R %s %S %T %u %w %x %X %y %Y %z %Z %% %+ %1 %2 %3 %4 %n %t'
FORMAT_='%F %T'
FORMAT='%a, %d %b %Y %T %z'

build: sft

bench: append.go write.go output_test.go
	go test -bench=. -benchmem $^

bencher: append_copy_test.go
	go version
	go test -bench=. $^

clean:
	rm -f sft append.go write.go append write

hyperfine: append write
	hyperfine './append' './write'

test: write
	./$<

# sft: main.go debug_development.go
# 	go build -tags sft_debug -o $@ $^

sft: main.go debug_release.go
	go build -o $@ $^

append: append.go output.go
	go build -o $@ $^

append.go: sft
	./sft -f appendTime -append -o $@ $(FORMAT)

write: write.go output.go
	go build -o $@ $^

write.go: sft
	./sft -f writeTime -o $@ $(FORMAT)

.PHONY: build bench clean hyperfine test
