# based on https://habr.com/ru/post/461467/
#
# for watch using watchexec from brew - github.com/watchexec/watchexec
#
.PHONY: all clean data_image help run
.DEFAULT_GOAL := help

PROJECTNAME=$(shell basename "$(PWD)")
CWD = $(shell pwd)
SERVICE := service
VIDEODIR := videodir
VIDEODIR_PID=/tmp/.$(VIDEODIR).pid

## videodir: Build binary
$(VIDEODIR): assets.go
	@-go build -i -o bin/$@ ./cmd/$@/main.go
	@echo end-build $@

## service: Build windows service
$(SERVICE):
	GOOS=windows GOARCH=386 go build -i -o bin/videodir_$@.exe ./cmd/$@

## clean: Clean build cache and remove bin directory
clean:
	go clean
	go clean -testcache
	rm -rf bin

# generate assets for static files
assets.go:
	@-go-bindata -pkg videodir -o videodir/assets.go -nocompress -nocompress -prefix static static/

## start: Start videodir with watch
start:
	@-bash -c "trap '$(MAKE) stop' EXIT; $(MAKE) watch"

stop:
	@echo stop
	@-touch $(VIDEODIR_PID)
	@-kill `cat $(VIDEODIR_PID)` 2> /dev/null || true
	@-rm $(VIDEODIR_PID)
	@sleep 1

run: stop
	@echo run
	@-$(CWD)/bin/$(VIDEODIR) & echo $$! > $(VIDEODIR_PID)

watch:
	@echo watch
	@-watchexec --exts go \
		-w cmd/ -w videodir/ -i videodir/assets.go \
		"make $(VIDEODIR) run"

## help: Show commands.
help: Makefile
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
