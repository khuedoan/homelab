.POSIX:

TAG = homelab-tools
PROJECT_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST)))/../)
WORKSPACE = ${HOME}/homelab

default: build run

build:
	@docker build . --tag ${TAG}

run:
	@mkdir -p ${WORKSPACE}
	@docker run \
		--rm \
		--interactive \
		--tty \
		--network host \
		--env "TERM=${TERM}" \
		--env "HOME=${HOME}" \
		--env "PROJECT_ROOT=${PROJECT_ROOT}" \
		--volume "${HOME}:${HOME}" \
		--volume "${PROJECT_ROOT}:${WORKSPACE}" \
		--volume "/var/run/docker.sock:/var/run/docker.sock" \
		--volume "/etc/passwd:/etc/passwd" \
		--user "$(shell id -u ${USER}):$(shell id -g ${USER})" \
		--group-add "$(shell getent group docker | cut -d ':' -f 3)" \
		--workdir "${WORKSPACE}" \
		${TAG}
