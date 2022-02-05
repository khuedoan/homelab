.POSIX:

TAG = homelab-tools
ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST)))/../)

default: build run

build:
	@docker build . --tag ${TAG}

run:
	@docker run \
		--rm \
		--interactive \
		--tty \
		--network host \
		--env "TERM=${TERM}" \
		--env "HOME=${HOME}" \
		--volume "${ROOT_DIR}:${HOME}" \
		--volume "/var/run/docker.sock:/var/run/docker.sock" \
		--volume "/etc/passwd:/etc/passwd" \
		--user "$(shell id -u ${USER}):$(shell id -g ${USER})" \
		--group-add "$(shell getent group docker | cut -d ':' -f 3)" \
		--workdir "${HOME}" \
		${TAG}
